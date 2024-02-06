package messageservice

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
	"github.com/MykolaSainiuk/schatgo/src/repo/messagerepo"
	"github.com/MykolaSainiuk/schatgo/src/service/chatservice"
)

type MessageService struct {
	messageRepo *messagerepo.MessageRepo

	chatService *chatservice.ChatService
}

func NewMessageService(srv types.IServer) *MessageService {
	return &MessageService{
		messageRepo: messagerepo.NewMessageRepo(srv.GetDB()),

		chatService: chatservice.NewChatService(srv),
	}
}

func (service *MessageService) NewMessage(ctx context.Context, chatId string, userId string, data *dto.NewMessageInputDto) (string, error) {
	_chatId, _ := primitive.ObjectIDFromHex(chatId)

	chat, err := service.chatService.GetChatById(ctx, _chatId)
	if err != nil || chat == nil {
		slog.Info("no chat found by such id")
		return "", errors.Join(cmnerr.ErrNotFoundEntity, err)
	}

	_userId, _ := primitive.ObjectIDFromHex(userId)
	newMessage := &model.Message{
		Text:      data.Text,
		Image:     data.Image,
		Sent:      false,
		Received:  false,
		System:    false,
		User:      _userId,
		Chat:      _chatId,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return service.messageRepo.SaveMessage(ctx, newMessage)
}

func (service *MessageService) GetAllMessages(ctx context.Context, chatID string) ([]model.Message, error) {
	return service.messageRepo.GetMessagesByChatID(ctx, chatID, types.PaginationParams{})
}

func (service *MessageService) GetMessagesPaginated(ctx context.Context, chatID string, pgParams types.PaginationParams) ([]model.Message, error) {
	return service.messageRepo.GetMessagesByChatID(ctx, chatID, pgParams)
}
