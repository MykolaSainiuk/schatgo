package chatservice

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/MykolaSainiuk/schatgo/src/api/dto"
	"github.com/MykolaSainiuk/schatgo/src/common/cmnerr"
	"github.com/MykolaSainiuk/schatgo/src/common/types"
	"github.com/MykolaSainiuk/schatgo/src/model"
	"github.com/MykolaSainiuk/schatgo/src/repo/chatrepo"
	"github.com/MykolaSainiuk/schatgo/src/service/userservice"
)

type ChatService struct {
	chatRepo *chatrepo.ChatRepo

	userService *userservice.UserService
}

func NewChatService(srv types.IServer) *ChatService {
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
	existingChat, err := service.chatRepo.GetExistingChat(ctx, _userId, anotherUser.ID)
	if err != nil && !errors.Is(err, cmnerr.ErrNotFoundEntity) {
		return nil, err
	}
	if existingChat != nil {
		return existingChat, nil
	}

	newUserItem := &model.Chat{
		Name:        data.ChatName,
		Muted:       false,
		IconUri:     "",
		Users:       []primitive.ObjectID{_userId, anotherUser.ID},
		LastMessage: primitive.NilObjectID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// trxSession, err := service.chatRepo.GetDB().StartTransaction()
	// if err != nil {
	// 	return nil, err
	// }
	// defer trxSession.EndSession(ctx)

	// // Step 1: Define the callback that specifies the sequence of operations to perform inside the transaction.
	// result, err := trxSession.WithTransaction(ctx, func(sessionCtx mongo.SessionContext) (interface{}, error) {
	// 	// Important: You must pass sessionCtx as the Context parameter to the operations for them to be executed in the transaction.
	// 	newChat, err1 := service.chatRepo.SaveChat(sessionCtx, newUserItem)
	// 	if err != nil {
	// 		slog.Info("error saving chat")
	// 		return nil, err1
	// 	}

	// 	err2 := service.userService.RegisterNewChat(sessionCtx, newChat.ID, userId, anotherUser.ID.Hex())

	// 	return newChat, errors.Join(err1, err2)
	// })

	// newChat, ok := result.(*model.Chat)
	// if !ok {
	// 	return nil, errors.Join(err, errors.New("cannot cast result (newChat) to *model.Chat"))
	// }

	newChat, err := service.chatRepo.SaveChat(ctx, newUserItem)
	if err != nil {
		slog.Info("error saving chat")
		return nil, err
	}

	err2 := service.userService.RegisterNewChat(ctx, newChat.ID, userId, anotherUser.ID.Hex())

	return newChat, errors.Join(err, err2)
}

func (service *ChatService) GetAllChats(ctx context.Context, userID string) ([]model.Chat, error) {
	return service.chatRepo.GetChatsByUserID(ctx, userID, types.PaginationParams{})
}

func (service *ChatService) GetChatsPaginated(ctx context.Context, userID string, pgParams types.PaginationParams) ([]model.Chat, error) {
	return service.chatRepo.GetChatsByUserID(ctx, userID, pgParams)
}

func (service *ChatService) GetChatById(ctx context.Context, chatID primitive.ObjectID) (*model.Chat, error) {
	return service.chatRepo.GetChatByID(ctx, chatID)
}

// func (service *ChatService) GetChat(ctx context.Context, chatID string) (*model.ChatPopulated, error) {
// 	return service.chatRepo.GetChatByIdPopulated(ctx, chatID)
// }
