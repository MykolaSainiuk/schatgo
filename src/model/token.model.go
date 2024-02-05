package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Token struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	UserId    primitive.ObjectID `json:"userId" bson:"userId"`
	Username  string             `json:"Username" bson:"username"`
	Type      string             `json:"type" bson:"type"`
	Encoded   string             `json:"encoded" bson:"encoded"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
}

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)
