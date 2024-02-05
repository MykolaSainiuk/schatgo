package chatrepo

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/MykolaSainiuk/schatgo/src/common/cmnerr"
	"github.com/MykolaSainiuk/schatgo/src/common/types"
	"github.com/MykolaSainiuk/schatgo/src/model"
)

type ChatRepo struct {
	name       string
	collection *mongo.Collection
	db         types.IDatabase
}

func NewChatRepo(db types.IDatabase) *ChatRepo {
	name := "chats"
	return &ChatRepo{
		name:       "chats",
		collection: db.GetCollection(name),
		db:         db,
	}
}

func (repo *ChatRepo) GetDB() types.IDatabase {
	return repo.db
}

func (repo *ChatRepo) GetExistingChat(ctx context.Context, userId primitive.ObjectID, anotherUserId primitive.ObjectID) (*model.Chat, error) {
	var chat *model.Chat
	var err error
	if err = repo.collection.FindOne(ctx, bson.D{{
		Key: "$or",
		Value: []bson.D{
			{{Key: "users", Value: []primitive.ObjectID{userId, anotherUserId}}},
			{{Key: "users", Value: []primitive.ObjectID{anotherUserId, userId}}},
		}},
	}).Decode(&chat); err != nil || chat == nil {
		slog.Error("cannot retrieve chat from chats collection", slog.Any("error", err.Error()))
		if errors.Is(err, mongo.ErrNoDocuments) || chat == nil {
			return nil, errors.Join(cmnerr.ErrNotFoundEntity, err)
		}
		return nil, err
	}

	return chat, err
}

func (repo *ChatRepo) SaveChat(ctx context.Context, data *model.Chat) (*model.Chat, error) {
	r, err := repo.collection.InsertOne(ctx, data)
	if err == nil {
		slog.Debug("saved chat", slog.String("ID", r.InsertedID.(primitive.ObjectID).String()))

		return repo.GetChatByID(ctx, r.InsertedID.(primitive.ObjectID))
	}

	errText := err.Error()
	if strings.Contains(errText, "duplicate key error collection") {
		return nil, errors.Join(cmnerr.ErrUniqueViolation, err)
	}

	slog.Error("cannot save chat into chats collection", slog.String("error", errText))
	return nil, err
}

func (repo *ChatRepo) GetChatByID(ctx context.Context, id primitive.ObjectID) (*model.Chat, error) {
	var chat *model.Chat
	var err error
	if err = repo.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&chat); err != nil || chat == nil {
		slog.Error("cannot retrieve chat from chats collection", slog.Any("error", err.Error()))
		if errors.Is(err, mongo.ErrNoDocuments) || chat == nil {
			return nil, errors.Join(cmnerr.ErrNotFoundEntity, err)
		}
		return nil, err
	}
	return chat, err
}

func (repo *ChatRepo) GetChatsByUserID(ctx context.Context, id string, params ...any) ([]model.Chat, error) {
	_id, _ := primitive.ObjectIDFromHex(id)

	match := bson.D{{Key: "$match", Value: bson.D{{
		Key: "users", Value: _id,
	}}}}

	pipelineStages := mongo.Pipeline{match}

	pgParam := params[0].(types.PaginationParams)
	if pgParam.Limit != 0 {
		limit := bson.D{{Key: "$limit", Value: pgParam.Limit}}
		pipelineStages = append(pipelineStages, limit)
	}
	if pgParam.Page != 0 {
		skip := bson.D{{Key: "$skip", Value: (pgParam.Page - 1) * pgParam.Limit}}
		pipelineStages = append(pipelineStages, skip)
	}

	var chats []model.Chat
	cursor, err := repo.collection.Aggregate(ctx, pipelineStages)
	if err != nil {
		slog.Error("cannot retrieve chat from chats collection", slog.Any("error", err.Error()))
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &chats); err != nil {
		slog.Error("cannot retrieve chat from chats collection", slog.Any("error", err.Error()))
		return nil, err
	}

	if len(chats) == 0 {
		return make([]model.Chat, 0), nil
	}
	return chats, err
}
