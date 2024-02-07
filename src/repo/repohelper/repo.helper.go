package repohelper

import (
	"github.com/MykolaSainiuk/schatgo/src/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func RawDocToUserModel(rawDoc map[string]any) *model.User {
	rcontacts, _ := rawDoc["contacts"].(primitive.A)
	rchats, _ := rawDoc["chats"].(primitive.A)

	contacts, chats := shrinkObjectsToItsIDs(rcontacts), shrinkObjectsToItsIDs(rchats)

	return &model.User{
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

func RawPlainDocToChatModel(rawDoc map[string]any) *model.Chat {
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

	return &model.Chat{
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

func RawDocToChatModel(rawDoc map[string]any) *model.Chat {
	rusers, _ := rawDoc["users"].(primitive.A)
	// rlm := rawDoc["lastMessage"]

	users := shrinkObjectsToItsIDs(rusers)
	// lastMessage := shrinkObjectsToItsIDs([]any{rlm})

	name, ok := rawDoc["name"].(string)
	if !ok {
		name = ""
	}
	iconUri, ok := rawDoc["iconUri"].(string)
	if !ok {
		iconUri = ""
	}

	return &model.Chat{
		ID:          rawDoc["_id"].(primitive.ObjectID),
		Name:        name,
		Muted:       rawDoc["muted"].(bool),
		IconUri:     iconUri,
		CreatedAt:   rawDoc["createdAt"].(primitive.DateTime).Time(),
		UpdatedAt:   rawDoc["updatedAt"].(primitive.DateTime).Time(),
		Users:       users,
		LastMessage: primitive.NilObjectID,
	}
}

func RawPlainDocToMessageModel(rawDoc map[string]any) *model.Message {
	if len(rawDoc) == 0 {
		return nil
	}
	
	image, ok := rawDoc["image"].(string)
	if !ok {
		image = ""
	}

	return &model.Message{
		ID:    rawDoc["_id"].(primitive.ObjectID),
		Text:  rawDoc["text"].(string),
		Image: image,

		Sent:     rawDoc["sent"].(bool),
		Received: rawDoc["received"].(bool),
		System:   rawDoc["system"].(bool),

		User: rawDoc["user"].(primitive.ObjectID),
		Chat: rawDoc["chat"].(primitive.ObjectID),

		CreatedAt: rawDoc["createdAt"].(primitive.DateTime).Time(),
		UpdatedAt: rawDoc["updatedAt"].(primitive.DateTime).Time(),
	}
}


func RawDocToMessageModel(rawDoc map[string]any) *model.Message {
	if len(rawDoc) == 0 {
		return nil
	}
	
	image, ok := rawDoc["image"].(string)
	if !ok {
		image = ""
	}

	return &model.Message{
		ID:    rawDoc["_id"].(primitive.ObjectID),
		Text:  rawDoc["text"].(string),
		Image: image,

		Sent:     rawDoc["sent"].(bool),
		Received: rawDoc["received"].(bool),
		System:   rawDoc["system"].(bool),

		User: primitive.NilObjectID,
		Chat: rawDoc["chat"].(primitive.ObjectID),

		CreatedAt: rawDoc["createdAt"].(primitive.DateTime).Time(),
		UpdatedAt: rawDoc["updatedAt"].(primitive.DateTime).Time(),
	}
}
