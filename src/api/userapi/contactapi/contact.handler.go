package contactapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"

	"github.com/MykolaSainiuk/schatgo/src/api/dto"
	"github.com/MykolaSainiuk/schatgo/src/common/cmnerr"
	"github.com/MykolaSainiuk/schatgo/src/common/httpexp"
	"github.com/MykolaSainiuk/schatgo/src/common/types"
	"github.com/MykolaSainiuk/schatgo/src/model"
	"github.com/MykolaSainiuk/schatgo/src/service/userservice"
)

type ContactHandler struct {
	userService *userservice.UserService
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
//	@Router			/api/user/contact/add [put]
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

	err := handler.userService.AddContact(ctx, userID, body.UserName)
	if err != nil {
		if errors.Is(err, cmnerr.ErrNotFoundEntity) {
			httpexp.From(err, "user not found", http.StatusNotFound).Reply(w)
			return
		}
		cmnerr.Reply500(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(cmnerr.EmptyJSONResponse)
}

// ListAllContacts method
//
//	@Summary		List all potential contacts
//	@Description	Unpaginated list of all potential contacts of User
//	@Tags			contact
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200		{array}		dto.UserInfoOutputDto
//	@Failure		404		{object}	httpexp.HttpExp	"Not found user"
//	@Router			/api/user/contact/list/all [get]
func (handler *ContactHandler) ListAllContacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(types.TokenPayload{}).(*types.TokenPayload).UserID

	contacts, err := handler.userService.GetAllUsers(ctx, userID)

	renderContacts(w, contacts, err)
}

// ListContactsPaginated method
//
//	@Summary		List contacts paginated
//	@Description	Paginated list of User contacts
//	@Tags			contact
//	@Security		BearerAuth
//	@Param			page	path	string					false	"page number"
//	@Param			limit	path	string					false	"page size"
//	@Produce		json
//	@Success		200		{array}		dto.UserInfoOutputDto
//	@Failure		404		{object}	httpexp.HttpExp	"Not found user"
//	@Router			/api/user/contact/list [get]
func (handler *ContactHandler) ListContactsPaginated(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(types.TokenPayload{}).(*types.TokenPayload).UserID

	page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	if err != nil || page < 0 || page > 100 {
		page = 1
	}
	limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32)
	if err != nil || limit < 0 || limit > 100 {
		limit = 10
	}

	contacts, err := handler.userService.GetContactsPaginated(ctx, userID, types.PaginationParams{
		Page:  int(page),
		Limit: int(limit),
	})

	renderContacts(w, contacts, err)
}

func renderContacts(w http.ResponseWriter, contacts []model.User, err error) {
	if err != nil {
		cmnerr.Reply500(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(contacts)
	w.Write(res)
}

const (
	MsgInvalidNewContactInput = "invalid input to add contact"
)
