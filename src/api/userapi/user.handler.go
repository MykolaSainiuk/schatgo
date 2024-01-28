package userapi

import (
	"encoding/json"
	"net/http"

	"github.com/MykolaSainiuk/schatgo/src/common/httpexp"
	"github.com/MykolaSainiuk/schatgo/src/common/types"
	"github.com/MykolaSainiuk/schatgo/src/server"
	"github.com/MykolaSainiuk/schatgo/src/service/userservice"
)

type UserHandler struct {
	UserService *userservice.UserService
}

func NewUserHandler(srv *server.Server) *UserHandler {
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
//	@Success		200	{object}	model.User
//	@Failure		404	{object}	httpexp.HttpExp	"Not found user"
//	@Router			/user/me [get]
func (handler *UserHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(types.TokenPayload{}).(*types.TokenPayload).UserID

	user, err := handler.UserService.GetUser(ctx, userID)
	if err != nil {
		httpexp.FromError(err, http.StatusNotFound).SetMessage("user not found").Reply(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		AvatarUri string `json:"avatarUri"`
	}{
		ID:        user.ID.Hex(),
		Name:      user.Name,
		AvatarUri: user.AvatarUri})
	w.Write(res)
}
