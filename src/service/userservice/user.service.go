package userservice

import (
	"context"

	"github.com/MykolaSainiuk/schatgo/src/common/types"
	"github.com/MykolaSainiuk/schatgo/src/model"
	"github.com/MykolaSainiuk/schatgo/src/repo/userrepo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserService struct {
	userRepo *userrepo.UserRepo
}

func NewUserService(srv types.IServer) *UserService {
	return &UserService{
		userRepo: userrepo.NewUserRepo(srv.GetDB()),
	}
}

func (service *UserService) GetUserById(ctx context.Context, userID string) (*model.User, error) {
	return service.userRepo.GetUserByID(ctx, userID)
}

func (service *UserService) GetUser(ctx context.Context, userID string) (*model.UserPopulated, error) {
	return service.userRepo.GetUserByIdPopulated(ctx, userID)
}

func (service *UserService) GetUserByName(ctx context.Context, name string) (*model.User, error) {
	return service.userRepo.GetUserByName(ctx, name)
}

func (service *UserService) AddContact(ctx context.Context, userID string, contactName string) error {
	return service.userRepo.AddContact(ctx, userID, contactName)
}

func (service *UserService) GetAllContacts(ctx context.Context, userID string) ([]model.User, error) {
	return service.userRepo.GetUserContactsByID(ctx, userID, types.PaginationParams{})
}

func (service *UserService) GetContactsPaginated(ctx context.Context, userID string, pgParams types.PaginationParams) ([]model.User, error) {
	return service.userRepo.GetUserContactsByID(ctx, userID, pgParams)
}

func (service *UserService) RegisterNewChat(ctx context.Context, chatID primitive.ObjectID, userID string, anotherUserID string) error {
	return service.userRepo.AddChatIdToUsers(ctx, chatID, userID, anotherUserID)
}
