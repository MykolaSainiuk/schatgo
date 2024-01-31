package types

import (
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/mongo"
)

type TokenPayload struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}

type IServer interface {
	GetRouter() chi.Router
	GetDB() IDatabase
	Shutdown()
	Run() <-chan struct{}
}

type IDatabase interface {
	GetCollection(string) *mongo.Collection
	Shutdown()
}

type PaginationParams struct {
	Page  int
	Limit int
}
