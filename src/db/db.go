package db

import (
	"context"
	"log/slog"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type DB struct {
	Database    *mongo.Database
	Client      *mongo.Client
	Name        string
	Context     context.Context
	CancelFn    context.CancelFunc
	Collections map[string]*mongo.Collection
}

const DB_CONN_TIMEOUT int = 10

func ConnectDB(dbName string) (*DB, error) {
	conn := options.Client().ApplyURI(os.Getenv("MONGO_URI"))
	// set db conn timeout try
	dbConnCtx, dbConnCancelFn := context.WithTimeout(context.Background(), time.Duration(DB_CONN_TIMEOUT)*time.Second)

	var err error
	var client *mongo.Client
	// simply establish connection
	if client, err = mongo.Connect(dbConnCtx, conn); err != nil {
		slog.Error("Cannot connect MongoDb instance", slog.Any("error", err.Error()))
		dbConnCancelFn()
		return nil, err
	}
	slog.Info("MongoDB client connected successfully...")

	// check access
	if err = client.Ping(dbConnCtx, readpref.Primary()); err != nil {
		slog.Error("Cannot reach primary read replica", slog.Any("error", err.Error()))
		dbConnCancelFn()
		return nil, err
	}
	slog.Info("MongoDB access checked successfully...")

	// create empty db if there is no such
	dbInst := client.Database(dbName)
	// "migrations"
	collections := RollUpAndGetCollections(dbInst, dbName)

	db := &DB{
		Database:    client.Database(dbName),
		Client:      client,
		Name:        dbName,
		Context:     dbConnCtx,
		CancelFn:    dbConnCancelFn,
		Collections: collections,
	}

	return db, nil
}

func RollUpAndGetCollections(db *mongo.Database, dbName string) map[string]*mongo.Collection {
	collections := make(map[string]*mongo.Collection)

	// init collections
	collections["users"] = db.Collection("users")
	collections["chats"] = db.Collection("chats")
	collections["messages"] = db.Collection("messages")
	collections["tokens"] = db.Collection("tokens")

	slog.Info("MongoDB metadata rolled up successfully...")

	return collections
}

func (db *DB) GetCollection(name string) *mongo.Collection {
	return db.Collections[name]
}
