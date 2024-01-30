package userapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/MykolaSainiuk/schatgo/src/api/dto"
	"github.com/MykolaSainiuk/schatgo/src/common/cmnerr"
	"github.com/MykolaSainiuk/schatgo/src/common/httpexp"
	"github.com/MykolaSainiuk/schatgo/src/common/types"
	"github.com/MykolaSainiuk/schatgo/src/service/userservice"
)

type UserHandler struct {
	UserService *userservice.UserService
}

func NewUserHandler(srv types.IServer) *UserHandler {
	userService := userservice.NewUserService(srv)
	return &UserHandler{userService}
}

// GetUserInfo method
//
//	@Summary		Get user info
//	@Description	ME endpoint
//	@Tags			user
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200		{object}	dto.UserInfoOutputDto	"User object"
//	@Failure		404		{object}	httpexp.HttpExp	"Not found user"
//	@Router			/user/me [get]
func (handler *UserHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(types.TokenPayload{}).(*types.TokenPayload).UserID

	user, err := handler.UserService.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, cmnerr.ErrNotFoundEntity) {
			httpexp.UserNotFoundExp.Reply(w)
			return
		}
		cmnerr.LogAndReply500(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(dto.UserInfoOutputDto{
		ID:        user.ID.Hex(),
		Name:      user.Name,
		AvatarUri: user.AvatarUri})
	w.Write(res)
}
