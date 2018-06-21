package main

// User object
type User struct {
	// Users username that will be saved in plain text format
	Username string `json: "username, omitempty"`
	// A password will saved as a hash string
	Password string `json: "password, omitempty"`
}
