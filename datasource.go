package main

// DataSource serves simple methods to access an underlying Data Store
type DataSource interface {
	FindAll(interface{}, string) ([]interface{}, error)
	FindOne(interface{}, string) (interface{}, error)
	Insert(interface{}, string) (interface{}, error)
	Update(interface{}, interface{}, string) error
	Remove(interface{}, string) error
	NewID() string
	Close()
}
