package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

// TODO create handles for login/register users and modify every endpoint handle so that they check basic authentication credentials

// viewPrescription will take of servicing prescriptions to the caller. The
// call would be made with the following format
// curl http://<hostname>:<port>/prescription/{prescription id}
// the return type will be a json formatted string
func viewPrescription(w http.ResponseWriter, r *http.Request, ds DataSource) {

	vars := mux.Vars(r)
	collection := "prescriptions"

	fmt.Printf("vars[id]: %s\n", vars["id"])

	prescriptionQuery := Prescription{ID: bson.ObjectIdHex(vars["id"])}

	result, err := ds.FindOne(prescriptionQuery, collection)

	if err != nil {
		log.Printf("Could not retrieve record for the following query %v\n", prescriptionQuery)
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "could not retreive record")
	} else {
		log.Printf("retrieved record %v\n", result)
		w.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(w)
		encoder.Encode(result)
	}

}

// updatePrescription will update a prescription in the database. This will be achived by calling the api as follows:
// curl --header "Content-Type: application/json" -X PUT http://<hostname>:<port>/prescription/{prescription id} \
// -d "{name: <name of prescription if want to update>, directions: <if want to update directions you would pass this>, time: <if want to update time when taking medication>}"
func updatePrescription(w http.ResponseWriter, r *http.Request, ds DataSource) {
	vars := mux.Vars(r)
	prescription := Prescription{}
	bodyReader := r.Body
	collection := "prescriptions"

	decoder := json.NewDecoder(bodyReader)
	err := decoder.Decode(&prescription)
	log.Printf("decoded prescription is as follows: %v", prescription)
	prescription.ID = bson.ObjectIdHex(vars["id"])
	err = ds.Update(prescription, collection)

	if err != nil {
		log.Printf("error happened when updating record in database")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error saving to the database")
	} else {
		w.WriteHeader(http.StatusOK)
	}

}

// CreatePrescription will server creating a new prescription, one created the prescription will be returned back to the user with an id
// curl --header "Content-Type: application/json" -X POST http://<hostname>:<port>/prescription \
// -d "{name: <name of prescription>, directions: <directions to follow when taking prescription>, time: <times to take perscription>}"
func createPrescription(w http.ResponseWriter, r *http.Request, ds DataSource) {
	prescription := Prescription{}
	bodyReader := r.Body
	collection := "prescriptions"

	decoder := json.NewDecoder(bodyReader)

	err := decoder.Decode(&prescription)

	prescription.ID = bson.ObjectId(ds.NewID())
	log.Printf("decoded prescription is as follows: %v", prescription)
	if err != nil {
		log.Printf("error happened while decoding body of request %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error parsing body of request")
	} else {
		record, err := ds.Insert(prescription, collection)
		_ = record
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "%f", err)
		} else {
			encoder := json.NewEncoder(w)
			err := encoder.Encode(&record)

			if err != nil {
				w.WriteHeader(http.StatusOK)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}

		}

	}

}

// deletePrescription will serve removing a prescription. Call will be as follows:
// curl -X DELETE http://<hostname>:<port>/prescription/{prescription id}
func deletePrescription(w http.ResponseWriter, r *http.Request, ds DataSource) {
	vars := mux.Vars(r)
	collection := "prescriptions"
	queryPrescription := Prescription{ID: bson.ObjectIdHex(vars["id"])}

	err := ds.Remove(queryPrescription, collection)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

// Will return a list of all prescription for this user
// curl http://<hostname>:<port>/prescriptions
func listPrescriptions(w http.ResponseWriter, r *http.Request, ds DataSource) {
	collection := "prescriptions"
	prescriptions, err := ds.FindAll(collection)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("%v", err)
	} else {
		w.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(w)
		encoder.Encode(prescriptions)
	}
}

func makeHandle(fn func(http.ResponseWriter, *http.Request, DataSource), ds DataSource) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, ds)
	}
}

func main() {
	var timeOut time.Duration
	url := "localhost"
	database := "prescriptions"
	ds, err := OpenMongoConnection(url, database)

	if err != nil {
		log.Fatalf("error %v while connecting to the datasource\n", err)
		return
	}
	log.Println("Connected to database")

	flag.DurationVar(&timeOut, "graceful-timeout", timeOut*15, "the duration for which the server gracefully shutsdown")
	flag.Parse()
	mx := mux.NewRouter()

	serv := &http.Server{
		Addr:         "0.0.0.0:8080",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      mx,
	}

	// TODO: create registration link and login link
	mx.HandleFunc("/prescription/{id}", makeHandle(viewPrescription, ds)).Methods("GET")
	mx.HandleFunc("/prescription/{id}", makeHandle(updatePrescription, ds)).Methods("PUT")
	mx.HandleFunc("/prescription", makeHandle(createPrescription, ds)).Methods("POST")
	mx.HandleFunc("/prescription/{id}", makeHandle(deletePrescription, ds)).Methods("DELETE")
	mx.HandleFunc("/prescriptions", makeHandle(listPrescriptions, ds)).Methods("GET")

	// TODO: make server listen over tls/ssl with basic authentication
	go func() {
		if err := serv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	killSignal := make(chan os.Signal, 1)

	<-killSignal

	ctx, cancel := context.WithTimeout(context.Background(), timeOut)

	defer cancel()

	serv.Shutdown(ctx)

	ds.Close()

	log.Println("Server got shutdown")

	os.Exit(0)
}
