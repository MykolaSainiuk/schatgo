package userapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/MykolaSainiuk/schatgo/src/api/dto"
	"github.com/MykolaSainiuk/schatgo/src/common/cmnerr"
	"github.com/MykolaSainiuk/schatgo/src/common/httpexp"
	"github.com/MykolaSainiuk/schatgo/src/common/types"
	"github.com/MykolaSainiuk/schatgo/src/service/userservice"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type UserHandler struct {
	userService *userservice.UserService
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

	user, err := handler.userService.GetUserInfo(ctx, userID)
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

// PublishKeysMethod method
//
//	@Summary		Publish keys
//	@Description	Publish signed pre keys
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.PublishKeysDto	true	"Public pre keys"
//	@Success		200		{object}	dto.PublishKeysDto		"Ok"
//	@Failure		422		{object}	httpexp.HttpExp					"Validation error"
//	@Router			/api/user/pre-keys [post]
func (handler *UserHandler) PublishSignedPreKeys(w http.ResponseWriter, r *http.Request) {
	var body dto.PublishKeysDto
	data, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(data, &body); err != nil {
		httpexp.From(err, MsgInvalidPreKeysInput, http.StatusUnprocessableEntity).Reply(w)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(&body); err != nil {
		validationErrs := cmnerr.GetValidationErrors(err)
		httpexp.From(err, MsgInvalidPreKeysInput, http.StatusUnprocessableEntity, validationErrs...).Reply(w)
		return
	}

	ctx := r.Context()
	userID := ctx.Value(types.TokenPayload{}).(*types.TokenPayload).UserID

	err := handler.userService.SetPreSignedKeys(ctx, userID, &body)
	if err != nil {
		cmnerr.Reply500(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(cmnerr.EmptyJSONResponse)
}

// FetchSignedPreKeys method
//
//	@Summary		Fetch user public keys
//	@Description	Return User public signed pre-keys
//	@Tags			user
//	@Produce		json
//	@Security		BearerAuth
//
// @Param			userId		path		string				true 	"User ID"
//
//	@Success		200			{object}	dto.PublishKeysDto			"Public pre-keys"
//	@Failure		404			{object}	httpexp.HttpExp				"Not found user"
//	@Router			/api/user/pre-keys/{userId} [get]
func (handler *UserHandler) FetchSignedPreKeys(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "userId")

	preKeys, err := handler.userService.GetUserPreKeys(ctx, userID)
	if err != nil {
		if errors.Is(err, cmnerr.ErrNotFoundEntity) {
			httpexp.From(err, "user not found", http.StatusNotFound).Reply(w)
			return
		}
		cmnerr.Reply500(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	if preKeys == nil { // also valid case for very first session
		w.Write(cmnerr.EmptyJSONResponse)
		return
	}
	res, _ := json.Marshal(preKeys)
	w.Write(res)
}

const (
	MsgInvalidPreKeysInput = "invalid public pre-keys payload"
)
