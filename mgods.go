package main

import (
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Mgods is a struct that allows provides Data source implementations
type Mgods struct {
	db      string
	session *mgo.Session
}

// MinimalRecord is a minimal database record
type MinimalRecord struct {
	ID bson.ObjectId
}

// NewID retrieves a new unique id for a record
func (mgods *Mgods) NewID() string {
	return string(bson.NewObjectId())
}

// FindAll returns all records in the data storage
func (mgods *Mgods) FindAll(query interface{}, collection string) ([]interface{}, error) {
	var result []interface{}
	db := mgods.session.DB(mgods.db)
	err := db.C(collection).Find(query.(bson.M)).All(&result)

	return result, err
}

// FindOne will find a result based on the passed query structure and return a result that match
func (mgods *Mgods) FindOne(query interface{}, collection string) (interface{}, error) {

	var result interface{}
	db := mgods.session.DB(mgods.db)

	// err := db.C(collection).Find(bson.M{"_id": query.ID}).One(&result)
	err := db.C(collection).Find(query.(bson.M)).One(&result)
	return result, err

}

// Insert will insert the passed record into the array, it will return a return with a unique id
func (mgods *Mgods) Insert(record interface{}, collection string) (interface{}, error) {

	db := mgods.session.DB(mgods.db)

	fmt.Printf("record to write to collection %s record: %v\n", collection, record)
	err := db.C(collection).Insert(&record)

	return record, err
}

// Update will update a record in the database. The client of this method will
// send the required fields to update with their value as a struct as well as a query struct
// that will be used to find the record(s) to udpate
func (mgods *Mgods) Update(query interface{}, change interface{}, collection string) error {
	db := mgods.session.DB(mgods.db)
	// change := bson.M{"$set": updateFields}
	// err := db.C(collection).Update(bson.M{"_id": updateFields.ID}, change)
	err := db.C(collection).Update(query.(bson.M), change.(bson.M))

	return err
}

// Remove will delete a record from the data source
func (mgods *Mgods) Remove(query interface{}, collection string) error {
	db := mgods.session.DB(mgods.db)
	err := db.C(collection).Remove(query.(bson.M))

	return err
}

// OpenMongoConnection will connect to a database and will return a handle that will be used to
// perform crud operations on the database. In the parameters the url of the database server is passed
// as well as the database to connect to.
func OpenMongoConnection(url, db string) (*Mgods, error) {

	var mgod Mgods
	session, err := mgo.Dial(url)

	if err != nil {
		return nil, err
	}

	session.SetMode(mgo.Monotonic, true)
	mgod = Mgods{db: db, session: session}
	return &mgod, nil
}

// Close will close the database connection
func (mgods *Mgods) Close() {
	mgods.session.Close()
}
