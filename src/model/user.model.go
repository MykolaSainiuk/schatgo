package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name      string             `json:"name,omitempty" bson:"name"`
	AvatarUri string             `json:"avatarUri,omitempty" bson:"avatarUri"`

	Hash string `json:"-" bson:"hash"`

	Contacts []primitive.ObjectID `json:"contacts" bson:"contacts"`
	Chats    []primitive.ObjectID `json:"chats" bson:"chats"`

	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
}

type UserPopulated struct {
	*User

	Contacts []*User `json:"contacts" bson:"contacts"`
	Chats    []*Chat `json:"chats" bson:"chats"`
}
