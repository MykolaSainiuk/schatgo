package db

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Database struct {
	Database    *mongo.Database
	Client      *mongo.Client
	Name        string
	Context     context.Context
	CancelFn    context.CancelFunc
	Collections map[string]*mongo.Collection
}

const DB_CONN_TIMEOUT int = 10

func ConnectDB(dbName string) (*Database, error) {
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
	slog.Info("MongoDB client connected successfully")

	// check access
	if err = client.Ping(dbConnCtx, readpref.Primary()); err != nil {
		slog.Error("Cannot reach primary read replica", slog.Any("error", err.Error()))
		dbConnCancelFn()
		return nil, err
	}
	slog.Info("MongoDB access checked successfully")

	// create empty db if there is no such
	dbInst := client.Database(dbName)

	// "migrations"
	collections, err := RollUpAndGetCollections(dbConnCtx, dbInst, dbName)
	if err != nil {
		slog.Error("Cannot roll up MongoDB metadata", slog.Any("error", err.Error()))
		dbConnCancelFn()
		return nil, err
	}

	db := &Database{
		Database:    dbInst,
		Client:      client,
		Name:        dbName,
		Context:     dbConnCtx,
		CancelFn:    dbConnCancelFn,
		Collections: collections,
	}

	return db, nil
}

func RollUpAndGetCollections(ctx context.Context, db *mongo.Database, dbName string) (map[string]*mongo.Collection, error) {
	collections := make(map[string]*mongo.Collection)

	// init collections
	collections["users"] = db.Collection("users")
	collections["chats"] = db.Collection("chats")
	collections["messages"] = db.Collection("messages")
	collections["tokens"] = db.Collection("tokens")

	// indices
	tokenIndices := db.Collection("tokens").Indexes()
	indexModel1 := mongo.IndexModel{
		Keys: bson.D{
			{Key: "userId", Value: 1},
			{Key: "type", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	_, err := tokenIndices.CreateOne(ctx, indexModel1)
	if err != nil {
		slog.Error("Cannot create index for tokens collection", slog.Any("error", err.Error()))
		return nil, err
	}

	accessTokenLifetime, err := strconv.Atoi(os.Getenv("ACCESS_TOKEN_EXPIRATION_SECONDS"))
	if err != nil {
		accessTokenLifetime = 3600
	}
	indexModel2 := mongo.IndexModel{
		Keys: bson.D{
			{Key: "encoded", Value: 1},
		},
		Options: options.Index().SetExpireAfterSeconds(int32(accessTokenLifetime)),
	}
	_, err = tokenIndices.CreateOne(ctx, indexModel2)
	if err != nil {
		slog.Error("Cannot create index for tokens collection", slog.Any("error", err.Error()))
		return nil, err
	}

	slog.Info("MongoDB metadata rolled up successfully")
	return collections, nil
}

func (db *Database) GetCollection(name string) *mongo.Collection {
	return db.Collections[name]
}

func (db *Database) Shutdown() {
	// drop DB connections
	if db.CancelFn != nil {
		db.CancelFn()
	}
	if db.Client != nil && db.Context != nil {
		if err := db.Client.Disconnect(db.Context); err != nil {
			slog.Error(err.Error())
		}
	}
}

func (db *Database) StartTransaction() (mongo.Session, error) {
	return db.Client.StartSession()
}
