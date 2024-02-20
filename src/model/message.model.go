package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID            primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Text          string             `json:"text" bson:"text"`
	EncryptedText string             `json:"encryptedText" bson:"encryptedText"`
	Image         string             `json:"image" bson:"image"`

	Sent     bool `json:"sent" bson:"sent"`
	Received bool `json:"received" bson:"received"`
	System   bool `json:"system" bson:"system"`

	User primitive.ObjectID `json:"user" bson:"user"`
	Chat primitive.ObjectID `json:"chat" bson:"chat"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

type MessagePopulated struct {
	*Message

	User *User `json:"user" bson:"user"`
}
