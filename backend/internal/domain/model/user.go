package model

// User represents a regular (non-admin) user in the system
type User struct {
	ID        string  `json:"id" bson:"_id"`
	Email     string  `json:"email" bson:"email"`
	Password  string  `json:"-" bson:"password"` // Never exposed in JSON
	FirstName string  `json:"firstName" bson:"firstName"`
	LastName  string  `json:"lastName" bson:"lastName"`
	Address   Address `json:"address" bson:"address"`
	Created   int64   `json:"created" bson:"created"`
	Updated   int64   `json:"updated" bson:"updated"`
}

// Address represents a user's address
type Address struct {
	Street  string `json:"street" bson:"street"`
	ZipCode string `json:"zipCode" bson:"zipCode"`
	City    string `json:"city" bson:"city"`
}
