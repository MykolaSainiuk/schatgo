package userrepo

import (
	"context"
	"log/slog"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/MykolaSainiuk/schatgo/src/common/cmnerr"
	"github.com/MykolaSainiuk/schatgo/src/common/types"
	"github.com/MykolaSainiuk/schatgo/src/model"
)

type UserRepo struct {
	name       string
	collection *mongo.Collection
}

func NewUserRepo(db types.IDatabase) *UserRepo {
	name := "users"
	return &UserRepo{
		name:       "users",
		collection: db.GetCollection(name),
	}
}

func (repo *UserRepo) SaveUser(ctx context.Context, newUser *model.User) (string, error) {
	r, err := repo.collection.InsertOne(ctx, newUser)
	if err == nil {
		slog.Info("saved user", slog.String("ID", r.InsertedID.(primitive.ObjectID).String()))
		return r.InsertedID.(primitive.ObjectID).Hex(), nil
	}

	errText := err.Error()
	if strings.Contains(errText, "duplicate key error collection") {
		return "", cmnerr.ErrUniqueViolation
	}

	slog.Error("cannot save user into users collection", slog.String("error", errText))
	return "", err
}

func (repo *UserRepo) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	_id, _ := primitive.ObjectIDFromHex(id)

	var user *model.User
	var err error
	if err = repo.collection.FindOne(ctx, bson.D{{Key: "_id", Value: _id}}).Decode(&user); err != nil {
		return nil, err
	}
	if user == nil {
		return nil, cmnerr.ErrNotFoundEntity
	}

	return user, err
}

func (repo *UserRepo) GetUsers(ctx context.Context, page int, limit int) (*[]model.User, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(int64(limit)).SetSkip(int64(page * limit))

	cursor, err := repo.collection.Find(ctx, bson.D{}, opts)
	if err != nil || cursor == nil {
		cursor.Close(ctx)
		slog.Error("cannot retrieve users from collection", slog.String("error", err.Error()))
		return nil, err
	}
	defer cursor.Close(ctx)

	users := make([]model.User, limit)
	if err = cursor.All(ctx, &users); err != nil {
		slog.Error("cannot cast collection from cursor to slice", slog.String("error", err.Error()))
		return nil, err
	}

	return &users, nil
}

func (repo *UserRepo) GetUserByName(ctx context.Context, name string) (*model.User, error) {
	var user *model.User
	if err := repo.collection.FindOne(ctx, bson.D{{Key: "name", Value: name}}).Decode(&user); err != nil || user == nil {
		if err != nil {
			slog.Error("cannot retrieve user", slog.String("error", err.Error()))
		}
		return nil, cmnerr.ErrNotFoundEntity
	}
	return user, nil
}

func (repo *UserRepo) UpdateUser(ctx context.Context, id primitive.ObjectID, keyValueMap map[string]any) error {
	setData := make(bson.D, len(keyValueMap))
	i := 0
	for Key, Value := range keyValueMap {
		setData[i] = primitive.E{Key: Key, Value: Value}
		i++
	}

	r, err := repo.collection.UpdateByID(ctx, id, bson.D{{
		Key:   "$set",
		Value: setData,
	}})

	return handleUpdateError(err, r.MatchedCount, id.String())
}

func handleUpdateError(err error, matchedCount int64, id string) error {
	if err != nil {
		slog.Error("cannot update user into users collection", slog.String("error", err.Error()))
		return err
	}
	if matchedCount == 0 {
		return cmnerr.ErrNotFoundEntity
	}

	slog.Info("updated user", slog.String("ID", id))
	return nil
}

func (repo *UserRepo) DeleteUser(ctx context.Context, id string) error {
	r, err := repo.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		slog.Error("cannot delete user from users collection", slog.String("error", err.Error()))
		return err
	}
	if r.DeletedCount == 0 {
		return cmnerr.ErrNotFoundEntity
	}

	slog.Info("deleted user", slog.String("ID", id))
	return err
}
