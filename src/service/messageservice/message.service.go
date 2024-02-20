package messageservice

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/MykolaSainiuk/schatgo/src/api/dto"
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

func (service *MessageService) NewMessage(ctx context.Context, chatId string, userId string, data *dto.NewMessageInputDto) (primitive.ObjectID, error) {
	_chatId, _ := primitive.ObjectIDFromHex(chatId)

	chat, err := service.chatService.GetChatByID(ctx, _chatId)
	if err != nil || chat == nil {
		slog.Info("no chat found by such id")
		return primitive.NilObjectID, err
	}

	_userId, _ := primitive.ObjectIDFromHex(userId)
	newMessage := &model.Message{
		Text:          data.Text,
		Image:         data.Image,
		EncryptedText: data.EncryptedText,
		Sent:          true,
		Received:      true,
		System:        false,
		User:          _userId,
		Chat:          _chatId,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	newMessageId, err := service.messageRepo.SaveMessage(ctx, newMessage)
	if err != nil || newMessageId == primitive.NilObjectID {
		return primitive.NilObjectID, err
	}

	err = service.chatService.SetLastMessage(ctx, chat.ID, newMessageId)

	return newMessageId, err
}

func (service *MessageService) GetAllMessages(ctx context.Context, chatID string) ([]model.MessagePopulated, error) {
	return service.messageRepo.GetMessagesByChatID(ctx, chatID, types.PaginationParams{})
}

func (service *MessageService) GetMessagesPaginated(ctx context.Context, chatID string, pgParams types.PaginationParams) ([]model.MessagePopulated, error) {
	return service.messageRepo.GetMessagesByChatID(ctx, chatID, pgParams)
}

func (service *MessageService) ClearChatMessages(ctx context.Context, chatID string) error {
	return service.messageRepo.RemoveAllMessagesByChatID(ctx, chatID)
}
