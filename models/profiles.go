package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Profile struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id,omitempty" bson:"user_id"`
	FirstName string             `json:"first_name,omitempty" bson:"first_name,omitempty"`
	LastName  string             `json:"last_name,omitempty" bson:"last_name,omitempty"`
	Phone     string             `json:"phone,omitempty" bson:"phone,omitempty"`
}
