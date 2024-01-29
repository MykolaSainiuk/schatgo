package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name      string             `json:"name,omitempty" bson:"name" binding:"required,omitempty"`
	AvatarUri string             `json:"avatar_uri,omitempty" bson:"avatar_uri" binding:"required"`
	Hash      string             `json:"-" bson:"hash" binding:"required,min=6"`
	Contacts  []User             `json:"contacts" bson:"contacts"`
	// Chats     []Chat             `json:"chats" bson:"chats"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}
