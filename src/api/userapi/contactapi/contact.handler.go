package contactapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/MykolaSainiuk/schatgo/src/api/dto"
	"github.com/MykolaSainiuk/schatgo/src/common/cmnerr"
	"github.com/MykolaSainiuk/schatgo/src/common/httpexp"
	"github.com/MykolaSainiuk/schatgo/src/common/types"
	"github.com/MykolaSainiuk/schatgo/src/service/userservice"
)

type ContactHandler struct {
	UserService *userservice.UserService
}

func NewContactHandler(srv types.IServer) *ContactHandler {
	userService := userservice.NewUserService(srv)
	return &ContactHandler{userService}
}

// AddContact method
//
//	@Summary		Add contact to User
//	@Description	Adding another user as a contact
//	@Tags			contact
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.AddContactInputDto	true	"New contact input"
//	@Success		200
//	@Failure		404		{object}	httpexp.HttpExp	"Not found user"
//	@Router			/user/contact/add [put]
func (handler *ContactHandler) AddContact(w http.ResponseWriter, r *http.Request) {
	var body dto.AddContactInputDto
	data, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(data, &body); err != nil {
		httpexp.From(err, MsgInvalidNewContactInput, http.StatusUnprocessableEntity).Reply(w)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(&body); err != nil {
		validationErrs := cmnerr.GetValidationErrors(err)
		httpexp.From(err, MsgInvalidNewContactInput, http.StatusUnprocessableEntity, validationErrs...).Reply(w)
		return
	}

	ctx := r.Context()
	userID := ctx.Value(types.TokenPayload{}).(*types.TokenPayload).UserID

	err := handler.UserService.AddContact(ctx, userID, body.UserName)
	if err != nil {
		if errors.Is(err, cmnerr.ErrNotFoundEntity) {
			httpexp.From(err, "user not found", http.StatusNotFound).Reply(w)
			return
		}
		cmnerr.Reply500(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(nil)
}

// ListAllContacts method
//
//	@Summary		List all contacts
//	@Description	Unpaginated list of all contacts of User
//	@Tags			contact
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200
//	@Failure		404		{object}	httpexp.HttpExp	"Not found user"
//	@Router			/user/contact/list [get]
func (handler *ContactHandler) ListAllContacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(types.TokenPayload{}).(*types.TokenPayload).UserID

	contacts, err := handler.UserService.GetAllContacts(ctx, userID)
	if err != nil {
		if errors.Is(err, cmnerr.ErrNotFoundEntity) {
			httpexp.From(err, "user not found", http.StatusNotFound).Reply(w)
			return
		}
		cmnerr.Reply500(w, err)
		return
	}

	users := make([]dto.UserInfoOutputDto, len(contacts))
	for i := range contacts {
		users[i] = dto.UserInfoOutputDto{
			ID:        contacts[i].ID.Hex(),
			Name:      contacts[i].Name,
			AvatarUri: contacts[i].AvatarUri,
		}
	}

	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(users)
	w.Write(res)
}

const (
	MsgInvalidNewContactInput = "invalid input to add contact"
)
