package main

// DataSource serves simple methods to access an underlying Data Store
type DataSource interface {
	FindAll(interface{}, interface{}, string) error
	FindOne(interface{}, interface{}, string) error
	Insert(interface{}, string) error
	Update(interface{}, interface{}, string) error
	Remove(interface{}, string) error
	NewID() string
	Close()
}
