package authservice

import (
	"context"
	"log/slog"
	"time"

	"github.com/MykolaSainiuk/schatgo/src/api/dto"
	"github.com/MykolaSainiuk/schatgo/src/common/cmnerr"
	"github.com/MykolaSainiuk/schatgo/src/helper/jwthelper"
	"github.com/MykolaSainiuk/schatgo/src/helper/pwdhelper"
	"github.com/MykolaSainiuk/schatgo/src/model"
	"github.com/MykolaSainiuk/schatgo/src/repo/userrepo"
	"github.com/MykolaSainiuk/schatgo/src/server"
)

type AuthService struct {
	userRepo *userrepo.UserRepo
}

func NewAuthService(srv *server.Server) *AuthService {
	return &AuthService{
		userRepo: userrepo.NewUserRepo(srv.GetDB()),
	}
}

func (service *AuthService) RegisterNewUser(ctx context.Context, dto *dto.RegisterInputDto) (string, error) {
	hash, err := pwdhelper.HashPassword(dto.Password)
	if err != nil {
		slog.Error("failed to generate hash", slog.Any("error", err))
		return "", cmnerr.ErrHashGeneration
	}

	newUser := &model.User{
		Name:      dto.Name,
		AvatarUri: dto.AvatarUri,

		Hash: hash,
		Contacts: make([]model.User, 0),

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	newUserId, err := service.userRepo.SaveUser(ctx, newUser)
	if err != nil {
		return "", err
	}
	return newUserId, nil
}

func (service *AuthService) LoginUser(ctx context.Context, name string, rawPassword string) (string, error) {
	user, err := service.userRepo.GetUserByName(ctx, name)
	if err != nil {
		slog.Info("no such user found by name")
		return "", cmnerr.ErrNotFoundEntity
	}

	if !pwdhelper.CheckPasswordHash(rawPassword, user.Hash) {
		slog.Info("bad password")
		return "", cmnerr.ErrHashMismatch
	}

	accessToken, err := jwthelper.GenerateToken(user.ID.Hex(), user.Name)
	if err != nil {
		slog.Error("failed to generate token", slog.Any("error", err))
		return "", cmnerr.ErrGenerateAccessToken
	}

	return accessToken, nil
}
