package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Text  string             `json:"text" bson:"text"`
	Image string             `json:"image" bson:"image"`

	Sent     bool `json:"sent" bson:"muted"`
	Received bool `json:"received" bson:"received"`
	System   bool `json:"system" bson:"system"`

	User primitive.ObjectID `json:"user" bson:"user"`
	Chat primitive.ObjectID `json:"chat" bson:"chat"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}
