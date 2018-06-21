package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

const (
	serverKey               = "server.key"
	serverCert              = "server.crt"
	prescriptionsCollection = "prescriptions"
	usersCollection         = "users"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		_, err := r.Cookie("username")
		path := r.RequestURI

		if err != nil && !strings.Contains(path, "/register") && !strings.Contains(path, "/login") {
			http.Error(w, "Unathorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// viewPrescription will take of servicing prescriptions to the caller. The
// call would be made with the following format
// curl http://<hostname>:<port>/prescription/{prescription id}
// the return type will be a json formatted string
func viewPrescription(w http.ResponseWriter, r *http.Request, ds DataSource) {

	vars := mux.Vars(r)
	cookie, err := r.Cookie("username")

	if err != nil {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	fmt.Printf("vars[id]: %s\n", vars["id"])

	query := bson.M{"_id": bson.ObjectIdHex(vars["id"]), "owner": cookie.Value}

	result, err := ds.FindOne(query, prescriptionsCollection)

	if err != nil {
		log.Printf("Could not retrieve record for the following query %v\n", query)
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
	cookie, err := r.Cookie("username")

	if err != nil {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(bodyReader)
	err = decoder.Decode(&prescription)
	log.Printf("decoded prescription is as follows: %v", prescription)
	prescription.ID = bson.ObjectIdHex(vars["id"])

	update := bson.M{"$set": bson.M{"time": prescription.Time, "name": prescription.Name, "directions": prescription.Directions,
		"owner": cookie.Value}}

	query := bson.M{"_id": prescription.ID}

	err = ds.Update(query, update, prescriptionsCollection)

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
	cookie, err := r.Cookie("username")

	if err != nil {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(bodyReader)

	err = decoder.Decode(&prescription)

	prescription.ID = bson.ObjectId(ds.NewID())
	log.Printf("decoded prescription is as follows: %v", prescription)
	if err != nil {
		log.Printf("error happened while decoding body of request %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Error parsing body of request")
	} else {
		query := bson.M{"_id": prescription.ID, "name": prescription.Name, "directions": prescription.Directions,
			"owner": cookie.Value, "time": prescription.Time}
		record, err := ds.Insert(query, prescriptionsCollection)
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

	cookie, err := r.Cookie("username")

	if err != nil {
		http.Error(w, fmt.Sprintf("%v", err), http.StatusUnauthorized)
		return
	}

	// queryPrescription := Prescription{ID: bson.ObjectIdHex(vars["id"])}
	query := bson.M{"_id": bson.ObjectIdHex(vars["id"]), "owner": cookie.Value}

	err = ds.Remove(query, prescriptionsCollection)

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
	cookie, err := r.Cookie("username")

	if err != nil {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	prescriptions, err := ds.FindAll(bson.M{"owner": cookie.Value}, collection)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("%v", err)
	} else {
		w.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(w)
		encoder.Encode(prescriptions)
	}
}

// checkCredentials will check a users credentials against the datasource and set a cookie if the credentials
// matched correctly
func checkCredentials(w http.ResponseWriter, r *http.Request, ds DataSource) {

	var creds string
	auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

	if len(auth) != 2 || auth[0] != "Basic" {
		http.Error(w, "authorization failed", http.StatusUnauthorized)
	}

	log.Printf("register user authentication credentials is as follows: %s\n", auth[1])
	byteCreds, err := base64.StdEncoding.DecodeString(auth[1])

	if err != nil {
		http.Error(w, "Error decoding authentication credentails", http.StatusInternalServerError)
		return
	}

	creds = string(byteCreds[:])

	log.Printf("credentails are as follows: %s\n", creds)

	credsArray := strings.SplitN(creds, ":", 2)
	hasher := sha256.New()
	hasher.Write([]byte(credsArray[1]))

	password := string(hasher.Sum(nil))

	user, err := ds.FindOne(bson.M{"username": creds[0]}, usersCollection)

	if err != nil {
		http.Error(w, "wrong username", http.StatusUnauthorized)
		return
	}

	if user.(User).Password != password {
		http.Error(w, "wrong password", http.StatusUnauthorized)
		return
	}

	expiration := time.Now().Add(time.Hour)
	http.SetCookie(w, &http.Cookie{Name: "username", Value: credsArray[0], Expires: expiration})
	w.WriteHeader(http.StatusOK)
}

// will create a new user in the datasource. will return a cookie after successful registration
func registerUser(w http.ResponseWriter, r *http.Request, ds DataSource) {
	var creds string
	auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

	if len(auth) != 2 || auth[0] != "Basic" {
		http.Error(w, "authorization failed", http.StatusUnauthorized)
	}

	log.Printf("register user authentication credentials is as follows: %s\n", auth[1])
	byteCreds, err := base64.StdEncoding.DecodeString(auth[1])

	if err != nil {
		http.Error(w, "Error decoding authentication credentials", http.StatusInternalServerError)
		return
	}

	creds = string(byteCreds[:])

	log.Printf("credentials are as follows: %s\n", creds)

	credsArray := strings.SplitN(creds, ":", 2)
	hasher := sha256.New()
	hasher.Write([]byte(credsArray[1]))

	user := User{Username: credsArray[0], Password: string(hasher.Sum(nil))}

	_, err = ds.Insert(user, usersCollection)

	if err != nil {
		http.Error(w, fmt.Sprintf("Could not create user\n%v", err), http.StatusInternalServerError)
		return
	}

	expiration := time.Now().Add(time.Hour)

	http.SetCookie(w, &http.Cookie{Name: "username", Value: credsArray[0], Expires: expiration})

	w.WriteHeader(http.StatusOK)
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
	mx.Use(loggingMiddleware)

	serv := &http.Server{
		Addr:         ":8081",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      mx,
	}

	// TODO: create registration link and login link
	mx.HandleFunc("/register", makeHandle(registerUser, ds)).Methods("POST")
	mx.HandleFunc("/login", makeHandle(checkCredentials, ds)).Methods("POST")
	mx.HandleFunc("/prescription/{id}", makeHandle(viewPrescription, ds)).Methods("GET")
	mx.HandleFunc("/prescription/{id}", makeHandle(updatePrescription, ds)).Methods("PUT")
	mx.HandleFunc("/prescription", makeHandle(createPrescription, ds)).Methods("POST")
	mx.HandleFunc("/prescription/{id}", makeHandle(deletePrescription, ds)).Methods("DELETE")
	mx.HandleFunc("/prescriptions", makeHandle(listPrescriptions, ds)).Methods("GET")

	// TODO: make server listen over tls/ssl with basic authentication
	go func() {
		if err := serv.ListenAndServeTLS(serverCert, serverKey); err != nil {
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
