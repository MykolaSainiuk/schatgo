package userrepo

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

	slog.Error("cannot save user into users collection", slog.String("error", errText))
	return "", err
}

func (repo *UserRepo) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	_id, _ := primitive.ObjectIDFromHex(id)

	var user *model.User
	var err error
	if err = repo.collection.FindOne(ctx, bson.D{{Key: "_id", Value: _id}}).Decode(&user); err != nil || user == nil {
		slog.Error("cannot retrieve user from users collection", slog.Any("error", err.Error()))
		if errors.Is(err, mongo.ErrNoDocuments) || user == nil {
			return nil, errors.Join(cmnerr.ErrNotFoundEntity, err)
		}
		return nil, err
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
		slog.Error("cannot retrieve user from users collection", slog.Any("error", err.Error()))
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &users); err != nil {
		slog.Error("cannot retrieve user from users collection", slog.Any("error", err.Error()))
		return nil, err
	}
	if len(users) == 0 {
		return nil, cmnerr.ErrNotFoundEntity
	}

	d, ok := users[0].(primitive.D)
	if !ok {
		return nil, errors.New("cannot cast user to UserPopulated")
	}
	user := d.Map()

	contacts := []model.User{}
	rawContacts, ok := user["contacts"].(primitive.A)
	if ok && len(rawContacts) > 0 {
		for _, v := range rawContacts {
			contacts = append(contacts, rawDocToUserModel(v.(primitive.D).Map()))
		}
	}
	chats := []model.Chat{}
	rawChats, ok := user["chats"].(primitive.A)
	if ok && len(rawChats) > 0 {
		for _, v := range rawChats {
			chats = append(chats, rawDocToChatModel(v.(primitive.D).Map()))
		}
	}

	return &model.UserPopulated{
		User:     rawDocToUserModel(user),
		Contacts: contacts,
		Chats:    chats,
	}, nil
}

func rawDocToUserModel(rawDoc map[string]any) model.User {
	rcontacts, _ := rawDoc["contacts"].(primitive.A)
	rchats, _ := rawDoc["chats"].(primitive.A)

	contacts, chats := shrinkObjectsToItsIDs(rcontacts), shrinkObjectsToItsIDs(rchats)

	return model.User{
		ID:        rawDoc["_id"].(primitive.ObjectID),
		Name:      rawDoc["name"].(string),
		AvatarUri: rawDoc["avatarUri"].(string),
		Hash:      rawDoc["hash"].(string),

		Contacts: contacts,
		Chats:    chats,

		CreatedAt: rawDoc["createdAt"].(primitive.DateTime).Time(),
		UpdatedAt: rawDoc["updatedAt"].(primitive.DateTime).Time(),
	}
}

func shrinkObjectsToItsIDs(objects primitive.A) []primitive.ObjectID {
	_ids := []primitive.ObjectID{}
	for _, v := range objects {
		_id, ok := v.(primitive.ObjectID)
		if !ok {
			_id, ok = v.(primitive.D).Map()["_id"].(primitive.ObjectID)
		}
		if ok {
			_ids = append(_ids, _id)
		}
	}
	return _ids
}

func rawDocToChatModel(rawDoc map[string]any) model.Chat {
	users := []primitive.ObjectID{}
	rco, ok := rawDoc["users"].(primitive.A)
	if ok {
		for _, v := range rco {
			users = append(users, v.(primitive.ObjectID))
		}
	}
	name, ok := rawDoc["name"].(string)
	if !ok {
		name = ""
	}
	iconUri, ok := rawDoc["iconUri"].(string)
	if !ok {
		iconUri = ""
	}
	lastMessage, ok := rawDoc["lastMessage"].(primitive.ObjectID)
	if !ok {
		lastMessage = primitive.NilObjectID
	}

	return model.Chat{
		ID:      rawDoc["_id"].(primitive.ObjectID),
		Name:    name,
		Muted:   rawDoc["muted"].(bool),
		IconUri: iconUri,

		Users:       users,
		LastMessage: lastMessage,

		CreatedAt: rawDoc["createdAt"].(primitive.DateTime).Time(),
		UpdatedAt: rawDoc["updatedAt"].(primitive.DateTime).Time(),
	}
}

func (repo *UserRepo) GetUsers(ctx context.Context, page int, limit int) (*[]model.User, error) {
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(int64(limit)).SetSkip(int64(page * limit))

	cursor, err := repo.collection.Find(ctx, bson.D{}, opts)
	if err != nil || cursor == nil {
		cursor.Close(ctx)
		slog.Error("cannot retrieve users from collection", slog.Any("error", err.Error()))
		return nil, err
	}
	defer cursor.Close(ctx)

	users := make([]model.User, limit)
	if err = cursor.All(ctx, &users); err != nil {
		slog.Error("cannot cast collection from cursor to slice", slog.Any("error", err.Error()))
		return nil, err
	}

	return &users, nil
}

func (repo *UserRepo) GetUserByName(ctx context.Context, name string) (*model.User, error) {
	records := repo.collection.FindOne(ctx, bson.D{{Key: "name", Value: name}}, options.FindOne().SetProjection(bson.D{
		// {Key: "contacts", Value: 0},
		// {Key: "chats", Value: 0},
	}))
	// TODO: defect if not omit "contacts" field

	var user *model.User
	if err := records.Decode(&user); err != nil || user == nil {
		if errors.Is(err, mongo.ErrNoDocuments) || user == nil {
			return nil, errors.Join(cmnerr.ErrNotFoundEntity, err)
		}
		slog.Error("cannot retrieve user", slog.Any("error", err.Error()))
		return nil, err
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
		slog.Error("cannot update user of users collection", slog.Any("error", err.Error()))
		return err
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
		slog.Error("cannot delete user from users collection", slog.Any("error", err.Error()))
		return err
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
		slog.Error("cannot retrieve user from users collection", slog.Any("error", err.Error()))
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &contacts); err != nil {
		slog.Error("cannot retrieve user from users collection", slog.Any("error", err.Error()))
		return nil, err
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
