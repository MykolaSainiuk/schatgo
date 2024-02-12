package messageapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/MykolaSainiuk/schatgo/src/api/dto"
	"github.com/MykolaSainiuk/schatgo/src/common/cmnerr"
	"github.com/MykolaSainiuk/schatgo/src/common/httpexp"
	"github.com/MykolaSainiuk/schatgo/src/common/types"
	"github.com/MykolaSainiuk/schatgo/src/model"
	"github.com/MykolaSainiuk/schatgo/src/service/messageservice"
)

type MessageHandler struct {
	messageService *messageservice.MessageService
}

func NewMessageHandler(srv types.IServer) *MessageHandler {
	messageService := messageservice.NewMessageService(srv)
	return &MessageHandler{messageService}
}

// NewMessage method
//
// @Summary			Write new message
// @Description		Add new message into the chat
// @Tags			message
// @Security		BearerAuth
// @Accept			json
// @Produce			json
// @Param       	chatId  path      	string  				true  "Chat ID"
// @Param			body	body		dto.NewMessageInputDto	true	"New contact input"
// @Success			201		{object}	dto.NewMessageOutputDto	"Created"
// @Failure			404		{object}	httpexp.HttpExp	"Not found user"
// @Router			/api/message/{chatId}/new [put]
func (handler *MessageHandler) NewMessage(w http.ResponseWriter, r *http.Request) {
	var body dto.NewMessageInputDto
	data, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(data, &body); err != nil {
		httpexp.From(err, MsgInvalidNewMessageInput, http.StatusUnprocessableEntity).Reply(w)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(&body); err != nil {
		validationErrs := cmnerr.GetValidationErrors(err)
		httpexp.From(err, MsgInvalidNewMessageInput, http.StatusUnprocessableEntity, validationErrs...).Reply(w)
		return
	}

	ctx := r.Context()
	userID := ctx.Value(types.TokenPayload{}).(*types.TokenPayload).UserID
	chatId := chi.URLParam(r, "chatId")

	newMessageID, err := handler.messageService.NewMessage(ctx, chatId, userID, &body)
	if err != nil || newMessageID == primitive.NilObjectID {
		if errors.Is(err, cmnerr.ErrNotFoundEntity) {
			httpexp.From(err, cmnerr.ErrNotFoundEntity.Error(), http.StatusNotFound).Reply(w)
			return
		}
		cmnerr.Reply500(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	res, _ := json.Marshal(dto.NewMessageOutputDto{Id: newMessageID.Hex()})
	w.Write(res)
}

// ListAllMessages method
//
//	@Summary		List all messages
//	@Description	Unpaginated list of all chat messages
//	@Tags			message
//	@Security		BearerAuth
//	@Produce		json
//	@Param        	chatId   path      	string  			true	"Chat ID"
//	@Success		200		{array}		dto.MessageExtendedOutputDto
//	@Router			/api/message/{chatId}/list/all [get]
func (handler *MessageHandler) ListAllMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chatId := chi.URLParam(r, "chatId")

	messages, err := handler.messageService.GetAllMessages(ctx, chatId)

	renderChats(w, messages, err)
}

// ListMessagesPaginated method
//
//	@Summary		List messages paginated
//	@Description	Paginated list of Chat messages
//	@Tags			message
//	@Security		BearerAuth
//	@Param        	chatId  path   		string  true  	"Chat ID"
//	@Param			page	path		string	false	"page number"
//	@Param			limit	path		string	false	"page size"
//	@Produce		json
//	@Success		200		{array}		dto.MessageExtendedOutputDto
//	@Router			/api/message/{chatId}/list [get]
func (handler *MessageHandler) ListMessagesPaginated(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chatId := chi.URLParam(r, "chatId")

	page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 32)
	if err != nil || page < 0 || page > 100 {
		page = 1
	}
	limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 32)
	if err != nil || limit < 0 || limit > 100 {
		limit = 10
	}

	messages, err := handler.messageService.GetMessagesPaginated(ctx, chatId, types.PaginationParams{
		Page:  int(page),
		Limit: int(limit),
	})

	renderChats(w, messages, err)
}

func renderChats(w http.ResponseWriter, messages []model.MessagePopulated, err error) {
	if err != nil {
		cmnerr.Reply500(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(messages)
	w.Write(res)
}

const (
	MsgInvalidNewMessageInput = "invalid input to write new message"
)
