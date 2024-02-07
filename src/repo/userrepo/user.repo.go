package userrepo

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
	"github.com/MykolaSainiuk/schatgo/src/repo/repohelper"
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
		slog.Debug("saved user", slog.String("ID", r.InsertedID.(primitive.ObjectID).String()))
		return r.InsertedID.(primitive.ObjectID).Hex(), nil
	}

	errText := err.Error()
	if strings.Contains(errText, "duplicate key error collection") {
		return "", errors.Join(cmnerr.ErrUniqueViolation, err)
	}

	return "", fmt.Errorf("cannot save user into users collection: %w", err)
}

func (repo *UserRepo) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	_id, _ := primitive.ObjectIDFromHex(id)

	var user *model.User
	var err error
	if err = repo.collection.FindOne(ctx, bson.D{{Key: "_id", Value: _id}}).Decode(&user); err != nil || user == nil {
		if errors.Is(err, mongo.ErrNoDocuments) || user == nil {
			return nil, errors.Join(cmnerr.ErrNotFoundEntity, err)
		}
		return nil, fmt.Errorf("cannot retrieve user from users collection: %w", err)
	}

	return user, err
}

func (repo *UserRepo) GetUserByIdPopulated(ctx context.Context, id string) (*model.UserPopulated, error) {
	_id, _ := primitive.ObjectIDFromHex(id)

	match := bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: _id}}}}

	ls1 := bson.D{
		{Key: "from", Value: "users"},
		{Key: "localField", Value: "contacts"},
		{Key: "foreignField", Value: "_id"},
		{Key: "as", Value: "contacts"},
	}
	lookup1 := bson.D{{Key: "$lookup", Value: ls1}}

	ls2 := bson.D{
		{Key: "from", Value: "chats"},
		{Key: "localField", Value: "chats"},
		{Key: "foreignField", Value: "_id"},
		{Key: "as", Value: "chats"},
	}
	lookup2 := bson.D{{Key: "$lookup", Value: ls2}}

	pipelineStages := mongo.Pipeline{match, lookup1, lookup2}

	var users []any
	cursor, err := repo.collection.Aggregate(ctx, pipelineStages)
	if err != nil {
		return nil, fmt.Errorf("cannot aggregate users collection: %w", err)
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("cannot decode users from cursor: %w", err)
	}
	if len(users) == 0 {
		return nil, cmnerr.ErrNotFoundEntity
	}

	d, _ := users[0].(primitive.D)
	user := d.Map()

	contacts := []*model.User{}
	rawContacts, ok := user["contacts"].(primitive.A)
	if ok && len(rawContacts) > 0 {
		for _, v := range rawContacts {
			contacts = append(contacts, repohelper.RawDocToUserModel(v.(primitive.D).Map()))
		}
	}
	chats := []*model.Chat{}
	rawChats, ok := user["chats"].(primitive.A)
	if ok && len(rawChats) > 0 {
		for _, v := range rawChats {
			chats = append(chats, repohelper.RawPlainDocToChatModel(v.(primitive.D).Map()))
		}
	}

	return &model.UserPopulated{
		User:     repohelper.RawDocToUserModel(user),
		Contacts: contacts,
		Chats:    chats,
	}, nil
}

func (repo *UserRepo) GetUsers(ctx context.Context, page int, limit int) (*[]model.User, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(int64(limit)).SetSkip(int64(page * limit))

	cursor, err := repo.collection.Find(ctx, bson.D{}, opts)
	if err != nil || cursor == nil {
		return nil, fmt.Errorf("cannot retrieve users from collection: %w", err)
	}
	defer cursor.Close(ctx)

	users := make([]model.User, limit)
	if err = cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("cannot decode users from cursor: %w", err)
	}

	return &users, nil
}

func (repo *UserRepo) GetUserByName(ctx context.Context, name string) (*model.User, error) {
	var user *model.User
	if err := repo.collection.FindOne(ctx, bson.D{{Key: "name", Value: name}}).Decode(&user); err != nil || user == nil {
		if errors.Is(err, mongo.ErrNoDocuments) || user == nil {
			return nil, errors.Join(cmnerr.ErrNotFoundEntity, err)
		}
		return nil, fmt.Errorf("cannot retrieve user from users collection: %w", err)
	}

	return user, nil
}

func (repo *UserRepo) AddContact(ctx context.Context, id string, name string) error {
	contactUser, err := repo.GetUserByName(ctx, name)
	if err != nil {
		return err
	}

	_id, _ := primitive.ObjectIDFromHex(id)
	r, err := repo.collection.UpdateByID(ctx, _id, bson.M{"$addToSet": bson.M{"contacts": contactUser.ID}})

	return handleUpdateError(err, r.MatchedCount, id)
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
		return fmt.Errorf("cannot update user of users collection: %w", err)
	}
	if matchedCount == 0 {
		return errors.Join(cmnerr.ErrNotFoundEntity, err)
	}

	slog.Debug("updated user", slog.String("ID", id))
	return nil
}

func (repo *UserRepo) DeleteUser(ctx context.Context, id string) error {
	r, err := repo.collection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return fmt.Errorf("cannot delete user from users collection: %w", err)
	}
	if r.DeletedCount == 0 {
		return errors.Join(cmnerr.ErrNotFoundEntity, err)
	}

	slog.Debug("deleted user", slog.String("ID", id))
	return err
}

func (repo *UserRepo) GetUserContactsByID(ctx context.Context, id string, params ...any) ([]model.User, error) {
	_id, _ := primitive.ObjectIDFromHex(id)

	match := bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: _id}}}}
	ls := bson.D{
		{Key: "from", Value: "users"},
		{Key: "localField", Value: "contacts"},
		{Key: "foreignField", Value: "_id"},
		{Key: "as", Value: "contacts1"},
	}
	lookup := bson.D{{Key: "$lookup", Value: ls}}
	unwind := bson.D{{Key: "$unwind", Value: "$contacts1"}}
	replaceRoot := bson.D{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$contacts1"}}}}

	pipelineStages := mongo.Pipeline{match, lookup, unwind, replaceRoot}

	pgParam := params[0].(types.PaginationParams)
	if pgParam.Limit != 0 {
		limit := bson.D{{Key: "$limit", Value: pgParam.Limit}}
		pipelineStages = append(pipelineStages, limit)
	}
	if pgParam.Page != 0 {
		skip := bson.D{{Key: "$skip", Value: (pgParam.Page - 1) * pgParam.Limit}}
		pipelineStages = append(pipelineStages, skip)
	}

	var contacts []model.User
	cursor, err := repo.collection.Aggregate(ctx, pipelineStages)
	if err != nil {
		return nil, fmt.Errorf("cannot aggregate contacts collection: %w", err)
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &contacts); err != nil {
		return nil, fmt.Errorf("cannot decode contacts from cursor: %w", err)
	}

	return contacts, err
}

func (repo *UserRepo) AddChatIdToUsers(ctx context.Context, chatId primitive.ObjectID, id1 string, id2 string) error {
	_id1, _ := primitive.ObjectIDFromHex(id1)
	r, err := repo.collection.UpdateByID(ctx, _id1, bson.M{"$addToSet": bson.M{"chats": chatId}})

	err = handleUpdateError(err, r.MatchedCount, id1)
	if err != nil {
		return err
	}

	_id2, _ := primitive.ObjectIDFromHex(id2)
	r, err = repo.collection.UpdateByID(ctx, _id2, bson.M{"$addToSet": bson.M{"chats": chatId}})

	return handleUpdateError(err, r.MatchedCount, id2)
}

func (repo *UserRepo) GetAllUsers(ctx context.Context, userID string) ([]model.User, error) {
	_userId, _ := primitive.ObjectIDFromHex(userID)

	cursor, err := repo.collection.Find(ctx, bson.D{{
		Key:   "_id",
		Value: bson.D{{Key: "$ne", Value: _userId}},
	}})
	if err != nil || cursor == nil {
		return nil, fmt.Errorf("cannot retrieve users from collection: %w", err)
	}
	defer cursor.Close(ctx)

	var users []model.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("cannot decode users from cursor: %w", err)
	}

	return users, nil
}
