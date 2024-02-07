package tokenrepo

import (
	"context"
	"errors"
	"fmt"
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
	if err != nil {
		return fmt.Errorf("cannot delete user tokens: %w", err)
	}
	return nil
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

	return "", fmt.Errorf("cannot save token into tokens collection: %w", err)
}

func (repo *TokenRepo) ExistToken(ctx context.Context, encoded *string) (bool, error) {
	record := repo.collection.FindOne(ctx, bson.D{{Key: "encoded", Value: *encoded}}, options.FindOne().SetProjection(bson.D{
		{Key: "_id", Value: 1},
	}))
	var token *model.Token
	if err := record.Decode(&token); err != nil || token == nil {
		if errors.Is(err, mongo.ErrNoDocuments) || token == nil {
			return false, errors.Join(cmnerr.ErrNotFoundEntity, err)
		}
		return false, fmt.Errorf("cannot retrieve token from tokens collection: %w", err)
	}

	return token != nil, nil
}
