package dto

import "go.mongodb.org/mongo-driver/bson/primitive"

// -- RegisterUser
// RegisterInputDto
type RegisterInputDto struct {
	Name      string `json:"name" validate:"required,min=2"`
	Password  string `json:"password" validate:"required,min=6"`
	AvatarUri string `json:"avatarUri" validate:"url|uri|base64url"`
}

// RegisterOutputDto
type RegisterOutputDto struct {
	UserId string `json:"id"`
}

// -- LoginUser
// LoginInputDto
type LoginInputDto struct {
	Name     string `json:"name" validate:"required,min=2"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginOutputDto
type LoginOutputDto struct {
	AccessToken string `json:"token"`
	// RefreshToken string `json:"refresh_token"`
}

// UserInfoOutputDto
type UserInfoOutputDto struct {
	ID        string `json:"_id"`
	Name      string `json:"name"`
	AvatarUri string `json:"avatarUri"`

	Contacts []primitive.ObjectID `json:"contacts"`
	Chats    []primitive.ObjectID `json:"chats"`

	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// UserInfoExtendedOutputDto
type UserInfoExtendedOutputDto struct {
	ID        string `json:"_id"`
	Name      string `json:"name"`
	AvatarUri string `json:"avatarUri"`

	Contacts []UserInfoExtendedOutputDto `json:"contacts"`
	Chats    []ChatOutputDto             `json:"chats"`

	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// AddContactInputDto
type AddContactInputDto struct {
	UserName string `json:"username" validate:"required,min=2"`
}

// AddChatInputDto
type AddChatInputDto struct {
	UserName string `json:"username" validate:"required,min=2"`
	ChatName string `json:"chatName"`
}

// ChatOutputDto
type ChatOutputDto struct {
	ID          string               `json:"_id"`
	Name        string               `json:"name"`
	IconUri     string               `json:"iconUri"`
	Muted       bool                 `json:"muted"`
	Users       []primitive.ObjectID `json:"users"`
	LastMessage primitive.ObjectID   `json:"lastMessage"`
	CreatedAt   string               `json:"createdAt"`
	UpdatedAt   string               `json:"updatedAt"`
}

// ChatExtendedOutputDto
type ChatExtendedOutputDto struct {
	ID          string              `json:"_id"`
	Name        string              `json:"name"`
	IconUri     string              `json:"iconUri"`
	Muted       bool                `json:"muted"`
	Users       []UserInfoOutputDto `json:"users"`
	LastMessage MessageOutputDto    `json:"lastMessage"`
	CreatedAt   string              `json:"createdAt"`
	UpdatedAt   string              `json:"updatedAt"`
}

// NewMessageInputDto
type NewMessageInputDto struct {
	Text  string `json:"text" validate:"required,min=1"`
	Image string `json:"image"`
	// Image string `json:"image" validate:"required_without=text,url|uri|base64url"`
}

// NewMessageOutputDto
type NewMessageOutputDto struct {
	Id string `json:"id"`
}

// MessageOutputDto
type MessageOutputDto struct {
	ID        string             `json:"_id"`
	Text      string             `json:"text"`
	Image     string             `json:"image"`
	Sent      bool               `json:"sent"`
	Received  bool               `json:"received"`
	System    bool               `json:"system"`
	User      primitive.ObjectID `json:"user"`
	Chat      primitive.ObjectID `json:"chat"`
	CreatedAt string             `json:"createdAt"`
	UpdatedAt string             `json:"updatedAt"`
}
