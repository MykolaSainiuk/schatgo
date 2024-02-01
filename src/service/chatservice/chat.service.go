package chatservice

import (
	"context"
	"log/slog"
	"time"

	"github.com/MykolaSainiuk/schatgo/src/api/dto"
	"github.com/MykolaSainiuk/schatgo/src/common/cmnerr"
	"github.com/MykolaSainiuk/schatgo/src/common/types"
	"github.com/MykolaSainiuk/schatgo/src/model"
	"github.com/MykolaSainiuk/schatgo/src/repo/chatrepo"
	"github.com/MykolaSainiuk/schatgo/src/service/userservice"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatService struct {
	chatRepo *chatrepo.ChatRepo

	userService *userservice.UserService
}

func NewUserService(srv types.IServer) *ChatService {
	return &ChatService{
		chatRepo: chatrepo.NewChatRepo(srv.GetDB()),

		userService: userservice.NewUserService(srv),
	}
}

func (service *ChatService) CreateChat(ctx context.Context, userId string, data *dto.AddChatInputDto) (*model.Chat, error) {
	anotherUser, err := service.userService.GetUserByName(ctx, data.UserName)
	if err != nil || anotherUser == nil {
		slog.Info("no such user found by name")
		return nil, cmnerr.ErrNotFoundEntity
	}

	_userId, _ := primitive.ObjectIDFromHex(userId)
	newUserItem := &model.Chat{
		Name:        data.ChatName,
		Muted:       false,
		IconUri:     "",
		Users:       []primitive.ObjectID{_userId, anotherUser.ID},
		LastMessage: primitive.NilObjectID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	newChat, err := service.chatRepo.SaveChat(ctx, newUserItem)
	if err != nil {
		slog.Info("error saving chat")
		return nil, err
	}

	err = service.userService.RegisterNewChat(ctx, newChat.ID, userId, anotherUser.ID.Hex())

	return newChat, err
}

func (service *ChatService) GetAllChats(ctx context.Context, userID string) ([]model.Chat, error) {
	return service.chatRepo.GetChatsByUserID(ctx, userID, types.PaginationParams{})
}

func (service *ChatService) GetChatsPaginated(ctx context.Context, userID string, pgParams types.PaginationParams) ([]model.Chat, error) {
	return service.chatRepo.GetChatsByUserID(ctx, userID, pgParams)
}
