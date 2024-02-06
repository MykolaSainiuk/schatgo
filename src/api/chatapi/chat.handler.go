package chatapi

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
	"github.com/MykolaSainiuk/schatgo/src/service/chatservice"
)

type ChatHandler struct {
	ChatService *chatservice.ChatService
}

func NewChatHandler(srv types.IServer) *ChatHandler {
	chatService := chatservice.NewChatService(srv)
	return &ChatHandler{chatService}
}

// NewChat method
//
//	@Summary		Create new chat
//	@Description	Establish new chat for two users
//	@Tags			chat
//	@Security		BearerAuth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.AddChatInputDto	true	"New chat input"
//	@Success		201
//	@Failure		404		{object}	httpexp.HttpExp	"Not found user"
//	@Router			/api/chat/new [put]
func (handler *ChatHandler) NewChat(w http.ResponseWriter, r *http.Request) {
	var body dto.AddChatInputDto
	data, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(data, &body); err != nil {
		httpexp.From(err, MsgInvalidNewChatInput, http.StatusUnprocessableEntity).Reply(w)
		return
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(&body); err != nil {
		validationErrs := cmnerr.GetValidationErrors(err)
		httpexp.From(err, MsgInvalidNewChatInput, http.StatusUnprocessableEntity, validationErrs...).Reply(w)
		return
	}

	ctx := r.Context()
	userID := ctx.Value(types.TokenPayload{}).(*types.TokenPayload).UserID

	newChat, err := handler.ChatService.CreateChat(ctx, userID, &body)
	if err != nil {
		if errors.Is(err, cmnerr.ErrNotFoundEntity) {
			httpexp.From(err, "user not found", http.StatusNotFound).Reply(w)
			return
		}
		cmnerr.Reply500(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	res, _ := json.Marshal(newChat)
	w.Write(res)
}

// ListAllChats method
//
//	@Summary		List all chats
//	@Description	Unpaginated list of all user chats
//	@Tags			chat
//	@Security		BearerAuth
//	@Produce		json
//	@Success		200		{array}		dto.ChatOutputDto
//	@Failure		404		{object}	httpexp.HttpExp	"Not found user"
//	@Router			/api/chat/list/all [get]
func (handler *ChatHandler) ListAllChats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := ctx.Value(types.TokenPayload{}).(*types.TokenPayload).UserID

	chats, err := handler.ChatService.GetAllChats(ctx, userID)

	renderChats(w, chats, err)
}

// ListChatsPaginated method
//
//	@Summary		List chats paginated
//	@Description	Paginated list of User chats
//	@Tags			chat
//	@Security		BearerAuth
//	@Param			page	path	string					false	"page number"
//	@Param			limit	path	string					false	"page size"
//	@Produce		json
//	@Success		200		{array}		dto.ChatOutputDto
//	@Failure		404		{object}	httpexp.HttpExp	"Not found user"
//	@Router			/api/chat/list [get]
func (handler *ChatHandler) ListChatsPaginated(w http.ResponseWriter, r *http.Request) {
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

	chats, err := handler.ChatService.GetChatsPaginated(ctx, userID, types.PaginationParams{
		Page:  int(page),
		Limit: int(limit),
	})

	renderChats(w, chats, err)
}

func renderChats(w http.ResponseWriter, chats []model.ChatPopulated, err error) {
	if err != nil {
		if errors.Is(err, cmnerr.ErrNotFoundEntity) {
			httpexp.From(err, "user not found", http.StatusNotFound).Reply(w)
			return
		}
		cmnerr.Reply500(w, err)
		return
	}

	// resChats := make([]dto.ChatOutputDto, len(chats))
	// for i := range chats {
	// 	resChats[i] = dto.ChatOutputDto{
	// 		ID:      chats[i].ID.Hex(),
	// 		Name:    chats[i].Name,
	// 		IconUri: chats[i].IconUri,
	// 		Muted:   chats[i].Muted,
	// 		Users:   chats[i].Users,
	// 		LastMessage: chats[i].LastMessage.Hex(),
	// 	}
	// }

	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(chats)
	w.Write(res)
}

const (
	MsgInvalidNewChatInput = "invalid input to create new chat"
)
