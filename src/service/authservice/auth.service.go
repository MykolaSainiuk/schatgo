package authservice

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/MykolaSainiuk/schatgo/src/api/dto"
	"github.com/MykolaSainiuk/schatgo/src/common/cmnerr"
	"github.com/MykolaSainiuk/schatgo/src/common/types"
	"github.com/MykolaSainiuk/schatgo/src/helper/jwthelper"
	"github.com/MykolaSainiuk/schatgo/src/helper/pwdhelper"
	"github.com/MykolaSainiuk/schatgo/src/model"
	"github.com/MykolaSainiuk/schatgo/src/repo/tokenrepo"
	"github.com/MykolaSainiuk/schatgo/src/repo/userrepo"
)

type AuthService struct {
	userRepo  *userrepo.UserRepo
	tokenRepo *tokenrepo.TokenRepo
}

func NewAuthService(srv types.IServer) *AuthService {
	return &AuthService{
		userRepo:  userrepo.NewUserRepo(srv.GetDB()),
		tokenRepo: tokenrepo.NewTokenRepo(srv.GetDB()),
	}
}

func (service *AuthService) RegisterNewUser(ctx context.Context, dto *dto.RegisterInputDto) (string, error) {
	hash, err := pwdhelper.HashPassword(dto.Password)
	if err != nil {
		slog.Error("failed to generate hash", slog.Any("error", err))
		return "", errors.Join(cmnerr.ErrHashGeneration, err)
	}

	newUser := &model.User{
		Name:      dto.Name,
		AvatarUri: dto.AvatarUri,

		Hash: hash,

		Contacts: make([]primitive.ObjectID, 0),
		Chats:    make([]primitive.ObjectID, 0),

		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	newUserId, err := service.userRepo.SaveUser(ctx, newUser)

	return newUserId, err
}

func (service *AuthService) LoginUser(ctx context.Context, name string, rawPassword string) (string, error) {
	user, err := service.userRepo.GetUserByName(ctx, name)
	if err != nil {
		slog.Info("no such user found by name")
		return "", err
	}

	if !pwdhelper.CheckPasswordHash(rawPassword, user.Hash) {
		slog.Info("bad password")
		return "", cmnerr.ErrHashMismatch
	}

	accessToken, err := jwthelper.GenerateToken(user.ID.Hex(), user.Name)
	if err != nil {
		slog.Error("failed to generate token", slog.Any("error", err))
		return "", errors.Join(cmnerr.ErrGenerateAccessToken, err)
	}

	err = service.tokenRepo.DeleteUserAccessTokens(ctx, user.ID)
	if err != nil {
		return "", err
	}
	_, err = service.tokenRepo.SaveToken(ctx, &model.Token{
		UserId:   user.ID,
		Username: user.Name,
		Type:     model.TokenTypeAccess,
		Encoded:  accessToken,
	})

	return accessToken, err
}
