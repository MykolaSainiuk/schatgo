package messagerepo

import (
	"context"
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/MykolaSainiuk/schatgo/src/common/types"
	"github.com/MykolaSainiuk/schatgo/src/model"
	"github.com/MykolaSainiuk/schatgo/src/repo/repohelper"
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

func (repo *MessageRepo) SaveMessage(ctx context.Context, newToken *model.Message) (primitive.ObjectID, error) {
	r, err := repo.collection.InsertOne(ctx, newToken)
	if err == nil {
		slog.Debug("saved message", slog.String("ID", r.InsertedID.(primitive.ObjectID).String()))
		return r.InsertedID.(primitive.ObjectID), nil
	}

	return primitive.NilObjectID, fmt.Errorf("cannot save message into messages collection: %w", err)
}

func (repo *MessageRepo) GetMessagesByChatID(ctx context.Context, id string, params ...any) ([]model.MessagePopulated, error) {
	_id, _ := primitive.ObjectIDFromHex(id)

	match := bson.D{{Key: "$match", Value: bson.D{{
		Key: "chat", Value: _id,
	}}}}
	pipelineStages := mongo.Pipeline{match}

	ls1 := bson.D{
		{Key: "from", Value: "users"},
		{Key: "localField", Value: "user"},
		{Key: "foreignField", Value: "_id"},
		{Key: "as", Value: "user"},
	}
	lookup1 := bson.D{{Key: "$lookup", Value: ls1}}
	unwind1 := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$user"}}}}
	sort := bson.D{{Key: "$sort", Value: bson.D{{Key: "createdAt", Value: -1}}}}

	pipelineStages = append(pipelineStages, lookup1, unwind1, sort)

	pgParam := params[0].(types.PaginationParams)
	if pgParam.Limit != 0 {
		limit := bson.D{{Key: "$limit", Value: pgParam.Limit}}
		pipelineStages = append(pipelineStages, limit)
	}
	if pgParam.Page != 0 {
		skip := bson.D{{Key: "$skip", Value: (pgParam.Page - 1) * pgParam.Limit}}
		pipelineStages = append(pipelineStages, skip)
	}

	var messages []any
	cursor, err := repo.collection.Aggregate(ctx, pipelineStages)
	if err != nil {
		return nil, fmt.Errorf("cannot aggregate from messages collection: %w", err)
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &messages); err != nil {
		return nil, fmt.Errorf("cannot decode messages from cursor: %w", err)
	}

	if len(messages) == 0 {
		return make([]model.MessagePopulated, 0), nil
	}

	messagesPopulated := make([]model.MessagePopulated, 0, len(messages))
	for i := range messages {
		d, _ := messages[i].(primitive.D)
		msg := d.Map()

		rawUser := msg["user"]

		messagesPopulated = append(messagesPopulated, model.MessagePopulated{
			Message: repohelper.RawDocToMessageModel(msg),
			User:    repohelper.RawDocToUserModel(rawUser.(primitive.D).Map()),
		})
	}

	return messagesPopulated, nil
}

func (repo *MessageRepo) RemoveAllMessagesByChatID(ctx context.Context, id string) error {
	_id, _ := primitive.ObjectIDFromHex(id)
	_, err := repo.collection.DeleteMany(ctx, bson.D{{Key: "chat", Value: _id}})

	return fmt.Errorf("cannot delete messages from messages collection: %w", err)
}
