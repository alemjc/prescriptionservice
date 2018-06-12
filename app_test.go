package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

const (
	hostname = "localhost"
	port     = "8080"
)

// TestPrescriptionCreation will test the creation of a prescription
func TestPrescriptionCreation(t *testing.T) {
	prescription := Prescription{Name: "Tylenol", Directions: "Use with lots of food", Time: "Every day"}
	b, err := json.Marshal(prescription)

	t.Logf("object to send %s\n", string(b))
	if err == nil {
		res, err := http.Post(fmt.Sprintf("http://%s:%s/prescription", hostname, port), "application/json", bytes.NewReader(b))

		if err == nil {

			var prescription Prescription
			decoder := json.NewDecoder(res.Body)
			decoder.Decode(&prescription)
			t.Logf("response received: %v. status %v", prescription, res.Status)
		} else {
			t.Errorf("error received when creating a new prescription %v", err)
		}
	} else {
		t.Errorf("error thrown when marshaling prescription object to create")
	}
}

// TsetUpdatingPrescription tests updating an already created prescription
func TestUpdatingPrescription(t *testing.T) {
	prescription := Prescription{Name: "Tylenol", Directions: "Drink after every meal", Time: "Every day"}
	client := http.Client{}
	b, err := json.Marshal(prescription)
	id := "5b1dc019b5b14812d96b43b6"

	t.Logf("object to send %s\n", string(b))

	if err == nil {
		req, err := http.NewRequest("PUT", fmt.Sprintf("http://%s:%s/prescription/%s", hostname, port, id), bytes.NewReader(b))
		req.Header.Add("Content-Type", "application/json")

		res, err := client.Do(req)

		if err != nil {
			t.Errorf("error received when updating a new prescription %v\n", err)
		} else {
			t.Logf("response status after updating the prescription %v", res.Status)
		}
	}
}

// TestDeletingPrescription will test deleting an already created prescription
func TestDeletingPrescription(t *testing.T) {
	id := "5b1dc019b5b14812d96b43b6"
	client := http.Client{}

	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://%s:%s/prescription/%s", hostname, port, id), nil)

	if err != nil {
		t.Errorf("error could not create request")
	} else {
		res, err := client.Do(req)

		if err != nil {
			t.Errorf("error received when removing prescription with id %s, error is %v\n", id, err)
		} else {
			t.Logf("Successfully removed prescription with id: %s status received %v\n", id, res.Status)
		}
	}
}

// TestPrescriptionRetreival will test retrieving an already created prescription
func TestPrescriptionRetreival(t *testing.T) {
	id := "5b1dc019b5b14812d96b43b6"

	res, err := http.Get(fmt.Sprintf("http://%s:%s/prescription/%s", hostname, port, id))

	if err != nil {
		t.Errorf("error received when sending request to retreive prescription")
	} else {
		var prescription Prescription
		decoder := json.NewDecoder(res.Body)
		decoder.Decode(&prescription)
		t.Logf("response received: %v. status %v", prescription, res.Status)
	}
}

// TestListPrescription will test retrieving all the prescriptions
func TestListPrescriptions(t *testing.T) {

	res, err := http.Get(fmt.Sprintf("http://%s:%s/prescription", hostname, port))

	if err != nil {
		t.Errorf("error received when sending request to retreive prescriptions")
	} else {
		var prescriptions []Prescription
		decoder := json.NewDecoder(res.Body)
		decoder.Decode(&prescriptions)
		t.Logf("response received: %v. status %v", prescriptions, res.Status)
	}
}
