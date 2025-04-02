package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// User represents a user in the system.
type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Username string             `bson:"username" json:"username"`
	Password string             `bson:"password" json:"-"` // Store hash, don't expose in JSON
}
