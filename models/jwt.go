package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type JWTToken struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Token string             `bson:"token"`
}
