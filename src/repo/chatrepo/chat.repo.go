package chatrepo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/MykolaSainiuk/schatgo/src/common/cmnerr"
	"github.com/MykolaSainiuk/schatgo/src/common/types"
	"github.com/MykolaSainiuk/schatgo/src/model"
	"github.com/MykolaSainiuk/schatgo/src/repo/repohelper"
)

type ChatRepo struct {
	name       string
	collection *mongo.Collection
}

func NewChatRepo(db types.IDatabase) *ChatRepo {
	name := "chats"
	return &ChatRepo{
		name:       "chats",
		collection: db.GetCollection(name),
	}
}

func (repo *ChatRepo) GetExistingChat(ctx context.Context, userId primitive.ObjectID, anotherUserId primitive.ObjectID) (*model.Chat, error) {
	var chat *model.Chat
	if err := repo.collection.FindOne(ctx, bson.D{{
		Key: "$or",
		Value: []bson.D{
			{{Key: "users", Value: []primitive.ObjectID{userId, anotherUserId}}},
			{{Key: "users", Value: []primitive.ObjectID{anotherUserId, userId}}},
		}},
	}).Decode(&chat); err != nil || chat == nil {
		if errors.Is(err, mongo.ErrNoDocuments) || chat == nil {
			return nil, errors.Join(cmnerr.ErrNotFoundEntity, err)
		}
		return nil, fmt.Errorf("cannot retrieve chat from chats collection: %w", err)
	}

	return chat, nil
}

func (repo *ChatRepo) SaveChat(ctx context.Context, data *model.Chat) (*model.Chat, error) {
	r, err := repo.collection.InsertOne(ctx, data)
	if err == nil {
		slog.Debug("saved chat", slog.String("ID", r.InsertedID.(primitive.ObjectID).String()))
		return repo.GetChatByID(ctx, r.InsertedID.(primitive.ObjectID))
	}

	return nil, fmt.Errorf("cannot save chat into chats collection: %w", err)
}

func (repo *ChatRepo) GetChatByID(ctx context.Context, id primitive.ObjectID) (*model.Chat, error) {
	var chat *model.Chat
	if err := repo.collection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&chat); err != nil || chat == nil {
		if errors.Is(err, mongo.ErrNoDocuments) || chat == nil {
			return nil, errors.Join(cmnerr.ErrNotFoundEntity, err)
		}
		return nil, fmt.Errorf("cannot retrieve chat from chats collection: %w", err)
	}

	return chat, nil
}

func (repo *ChatRepo) GetChatsByUserID(ctx context.Context, id string, params ...any) ([]model.ChatPopulated, error) {
	_id, _ := primitive.ObjectIDFromHex(id)

	match := bson.D{{Key: "$match", Value: bson.D{{
		Key: "users", Value: _id,
	}}}}

	ls1 := bson.D{
		{Key: "from", Value: "users"},
		{Key: "localField", Value: "users"},
		{Key: "foreignField", Value: "_id"},
		{Key: "as", Value: "users"},
	}
	lookup1 := bson.D{{Key: "$lookup", Value: ls1}}

	ls2 := bson.D{
		{Key: "from", Value: "messages"},
		{Key: "localField", Value: "lastMessage"},
		{Key: "foreignField", Value: "_id"},
		{Key: "as", Value: "lastMessage"},
	}
	lookup2 := bson.D{{Key: "$lookup", Value: ls2}}
	// {Key: "preserveNullAndEmptyArrays", Value: true}
	unwind2 := bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$lastMessage"}}}}

	pipelineStages := mongo.Pipeline{match, lookup1, lookup2, unwind2}

	pgParam := params[0].(types.PaginationParams)
	if pgParam.Limit != 0 {
		limit := bson.D{{Key: "$limit", Value: pgParam.Limit}}
		pipelineStages = append(pipelineStages, limit)
	}
	if pgParam.Page != 0 {
		skip := bson.D{{Key: "$skip", Value: (pgParam.Page - 1) * pgParam.Limit}}
		pipelineStages = append(pipelineStages, skip)
	}

	var chats []any
	cursor, err := repo.collection.Aggregate(ctx, pipelineStages)
	if err != nil {
		return nil, fmt.Errorf("cannot aggregate from chats collection: %w", err)
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &chats); err != nil {
		return nil, fmt.Errorf("cannot decode chats from cursor: %w", err)
	}

	if len(chats) == 0 {
		return make([]model.ChatPopulated, 0), nil
	}

	chatsPopulated := make([]model.ChatPopulated, 0, len(chats))
	for i := range chats {
		d, _ := chats[i].(primitive.D)
		chat := d.Map()

		users := []*model.User{}
		rawUsers, ok := chat["users"].(primitive.A)
		if ok && len(rawUsers) > 0 {
			for _, v := range rawUsers {
				users = append(users, repohelper.RawDocToUserModel(v.(primitive.D).Map()))
			}
		}

		rawLM := chat["lastMessage"]

		chatsPopulated = append(chatsPopulated, model.ChatPopulated{
			Chat:        repohelper.RawDocToChatModel(chat),
			Users:       users,
			LastMessage: repohelper.RawPlainDocToMessageModel(rawLM.(primitive.D).Map()),
		})
	}

	return chatsPopulated, nil
}

func (repo *ChatRepo) SetLastMessage(ctx context.Context, chatID, messageID primitive.ObjectID) error {
	r, err := repo.collection.UpdateByID(ctx, chatID, bson.D{{
		Key:   "$set",
		Value: primitive.D{{Key: "lastMessage", Value: messageID}},
	}})
	if err != nil {
		return fmt.Errorf("cannot update chat of chats collection: %w", err)
	}
	if r.MatchedCount == 0 {
		return errors.Join(cmnerr.ErrNotFoundEntity, err)
	}

	slog.Debug("updated last message for chat", slog.String("ID", chatID.Hex()))
	return nil
}
