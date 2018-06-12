package main

import (
	"gopkg.in/mgo.v2/bson"
)

// Prescription is a structure that details a user prescription that would be saved in the environment
type Prescription struct {
	Directions string        `bson:"directions" json: "directions, omitempty"`
	Time       string        `bson:"time" json: "time, omitempty"`
	Name       string        `bson:"name" json: "name"`
	ID         bson.ObjectId `bson:"_id,omitempty" json:"id, omitempty"`
}
