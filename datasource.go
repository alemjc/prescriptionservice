package main

// DataSource serves simple methods to access an underlying Data Store
type DataSource interface {
	FindAll(string) ([]Prescription, error)
	FindOne(Prescription, string) (Prescription, error)
	Insert(Prescription, string) (Prescription, error)
	Update(Prescription, string) error
	Remove(Prescription, string) error
	NewID() string
	Close()
}
