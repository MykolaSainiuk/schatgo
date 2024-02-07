package authapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/MykolaSainiuk/schatgo/src/api/dto"
	"github.com/MykolaSainiuk/schatgo/src/common/cmnerr"
	"github.com/MykolaSainiuk/schatgo/src/common/httpexp"
	"github.com/MykolaSainiuk/schatgo/src/common/types"
	"github.com/MykolaSainiuk/schatgo/src/service/authservice"
)

type AuthHandler struct {
	authService *authservice.AuthService
}

func NewAuthHandler(srv types.IServer) *AuthHandler {
	authService := authservice.NewAuthService(srv)
	return &AuthHandler{authService}
}

// RegisterUser method
//
//	@Summary		Register user
//	@Description	Register user with name & avatar
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.RegisterInputDto	true	"Registration form"
//	@Success		201		{object}	dto.RegisterOutputDto	"Created"
//	@Failure		422		{object}	httpexp.HttpExp	"Validation error"
//	@Router			/auth/register [post]
func (handler *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var body dto.RegisterInputDto
	data, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(data, &body); err != nil {
		httpexp.From(err, MsgInvalidRegisterInput, http.StatusUnprocessableEntity).Reply(w)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(&body); err != nil {
		validationErrs := cmnerr.GetValidationErrors(err)
		httpexp.From(err, MsgInvalidRegisterInput, http.StatusUnprocessableEntity, validationErrs...).Reply(w)
		return
	}
	if bytes.Count([]byte(body.AvatarUri), nil) > 2047 {
		httpexp.From(errors.New("avatarUri is too long"), MsgInvalidRegisterInput, http.StatusUnprocessableEntity).Reply(w)
		return
	}

	ctx := r.Context()
	userId, err := handler.authService.RegisterNewUser(ctx, &body)
	if err != nil {
		if errors.Is(err, cmnerr.ErrUniqueViolation) {
			httpexp.From(err, "such name is already occupied", http.StatusUnprocessableEntity).Reply(w)
			return
		}
		// if errors.Is(err, cmnerr.ErrHashGeneration) {
		// }
		cmnerr.Reply500(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	res, _ := json.Marshal(dto.RegisterOutputDto{UserId: userId})
	w.Write(res)
}

// LoginUser method
//
//	@Summary		Login user
//	@Description	Make access token for user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.LoginInputDto	true	"Registration form"
//	@Success		200		{object}	dto.LoginOutputDto
//	@Failure		422		{object}	httpexp.HttpExp	"Validation error"
//	@Router			/auth/login [post]
func (handler *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var body dto.LoginInputDto
	data, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(data, &body); err != nil {
		httpexp.From(err, MsgFailedToLogin, http.StatusUnauthorized).Reply(w)
		return
	}

	validate := validator.New()
	if err := validate.Struct(body); err != nil {
		validationErrs := cmnerr.GetValidationErrors(err)
		httpexp.From(err, MsgFailedToLogin, http.StatusUnauthorized, validationErrs...).Reply(w)
		return
	}
	// FromError(err, http.StatusUnauthorized).SetNewMessage(failedToLoginMsg) - bcz always oblivious about reasons

	ctx := r.Context()
	accessToken, err := handler.authService.LoginUser(ctx, body.Name, body.Password)
	if err != nil {
		if errors.Is(err, cmnerr.ErrNotFoundEntity) || errors.Is(err, cmnerr.ErrHashMismatch) || errors.Is(err, cmnerr.ErrGenerateAccessToken) {
			httpexp.From(err, MsgFailedToLogin, http.StatusUnauthorized).Reply(w)
			return
		}
		cmnerr.Reply500(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	rawRes := r.URL.Query().Get("raw")
	if rawRes == "true" {
		w.Write([]byte(accessToken))
	} else {
		res, _ := json.Marshal(dto.LoginOutputDto{AccessToken: accessToken})
		w.Write(res)
	}
}

const (
	MsgFailedToLogin        = "failed to login user"
	MsgInvalidRegisterInput = "invalid input to register user"
)
