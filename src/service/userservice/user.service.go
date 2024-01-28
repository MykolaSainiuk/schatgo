package userservice

import (
	"context"

	"github.com/MykolaSainiuk/schatgo/src/model"
	"github.com/MykolaSainiuk/schatgo/src/repo/userrepo"
	"github.com/MykolaSainiuk/schatgo/src/server"
)

type UserService struct {
	userRepo *userrepo.UserRepo
}

func NewUserService(srv *server.Server) *UserService {
	return &UserService{
		userRepo: userrepo.NewUserRepo(srv.GetDB()),
	}
}

func (service *UserService) GetUser(ctx context.Context, userID string) (*model.User, error) {
	return service.userRepo.GetUserByID(ctx, userID)
}
