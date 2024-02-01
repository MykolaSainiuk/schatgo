package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Chat struct {
	ID      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name    string             `json:"name" bson:"name"`
	Muted   bool               `json:"muted" bson:"muted"`
	IconUri string             `json:"iconUri" bson:"iconUri"`

	Users       []primitive.ObjectID `json:"users" bson:"users"`
	LastMessage primitive.ObjectID   `json:"lastMessage" bson:"lastMessage"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}
