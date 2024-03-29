package userapi

import (
	"encoding/json"
	"errors"
	"net/http"

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
//	@Success		200		{object}	dto.UserInfoExtendedOutputDto	"User object extended"
//	@Failure		404		{object}	httpexp.HttpExp	"Not found user"
//	@Router			/api/user/me [get]
func (handler *UserHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(types.TokenPayload{}).(*types.TokenPayload).UserID

	user, err := handler.UserService.GetUserInfo(ctx, userID)
	if err != nil {
		if errors.Is(err, cmnerr.ErrNotFoundEntity) {
			httpexp.From(err, "user not found", http.StatusNotFound).Reply(w)
			return
		}
		cmnerr.Reply500(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(user)
	w.Write(res)
}
