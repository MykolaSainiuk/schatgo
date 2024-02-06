package messagerepo

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

type MessageRepo struct {
	name       string
	collection *mongo.Collection
}

func NewMessageRepo(db types.IDatabase) *MessageRepo {
	name := "messages"
	return &MessageRepo{
		name:       "messages",
		collection: db.GetCollection(name),
	}
}

func (repo *MessageRepo) SaveMessage(ctx context.Context, newToken *model.Message) (string, error) {
	r, err := repo.collection.InsertOne(ctx, newToken)
	if err == nil {
		slog.Debug("saved message", slog.String("ID", r.InsertedID.(primitive.ObjectID).String()))
		return r.InsertedID.(primitive.ObjectID).Hex(), nil
	}

	errText := err.Error()
	if strings.Contains(errText, "duplicate key error collection") {
		return "", errors.Join(cmnerr.ErrUniqueViolation, err)
	}

	slog.Error("cannot save message into messages collection", slog.String("error", errText))
	return "", err
}

func (repo *MessageRepo) GetMessagesByChatID(ctx context.Context, id string, params ...any) ([]model.Message, error) {
	_id, _ := primitive.ObjectIDFromHex(id)

	match := bson.D{{Key: "$match", Value: bson.D{{
		Key: "chat", Value: _id,
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

	var messages []model.Message
	cursor, err := repo.collection.Aggregate(ctx, pipelineStages)
	if err != nil {
		slog.Error("cannot retrieve message from messages collection", slog.Any("error", err.Error()))
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &messages); err != nil {
		slog.Error("cannot retrieve message from messages collection", slog.Any("error", err.Error()))
		return nil, err
	}

	if len(messages) == 0 {
		return make([]model.Message, 0), nil
	}
	return messages, err
}
