package tokenrepo

import (
	"context"
	"errors"
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

type TokenRepo struct {
	name       string
	collection *mongo.Collection
}

func NewTokenRepo(db types.IDatabase) *TokenRepo {
	name := "tokens"
	return &TokenRepo{
		name:       "tokens",
		collection: db.GetCollection(name),
	}
}

func (repo *TokenRepo) DeleteUserAccessTokens(ctx context.Context, userId primitive.ObjectID) error {
	_, err := repo.collection.DeleteMany(ctx, bson.M{
		"userId": userId,
		"type":   model.TokenTypeAccess,
	})
	return err
}

func (repo *TokenRepo) SaveToken(ctx context.Context, newToken *model.Token) (string, error) {
	r, err := repo.collection.InsertOne(ctx, newToken)
	if err == nil {
		slog.Debug("saved token", slog.String("ID", r.InsertedID.(primitive.ObjectID).String()))
		return r.InsertedID.(primitive.ObjectID).Hex(), nil
	}

	errText := err.Error()
	if strings.Contains(errText, "duplicate key error collection") {
		return "", errors.Join(cmnerr.ErrUniqueViolation, err)
	}

	slog.Error("cannot save token into tokens collection", slog.String("error", errText))
	return "", err
}

func (repo *TokenRepo) ExistToken(ctx context.Context, encoded *string) (bool, error) {
	var token *model.Token
	var err error
	record := repo.collection.FindOne(ctx, bson.D{{Key: "encoded", Value: *encoded}}, options.FindOne().SetProjection(bson.D{
		{Key: "_id", Value: 1},
	}))
	if err = record.Decode(&token); err != nil || token == nil {
		slog.Error("cannot retrieve token from tokens collection", slog.Any("error", err.Error()))
		if errors.Is(err, mongo.ErrNoDocuments) || token == nil {
			return false, errors.Join(cmnerr.ErrNotFoundEntity, err)
		}
		return false, err
	}

	return token != nil, err
}
