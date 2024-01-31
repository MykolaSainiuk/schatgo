package httpexp

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

// InvalidResponseDto
//
//	@Description	common 422 response
type HttpExp struct {
	Error      error `json:"-"`
	StatusCode int   `json:"-"`

	PublicMessage string   `json:"message"`
	Details       []string `json:"details,omitempty"`
}

func From(err error, msg string, code int, details ...string) *HttpExp {
	return &HttpExp{
		Error:         err,
		StatusCode:    code,
		PublicMessage: msg,
		Details:       details,
	}
}

func (e *HttpExp) SetMessage(msg string) *HttpExp {
	e.PublicMessage = msg
	return e
}

func (e *HttpExp) Reply(w http.ResponseWriter) {
	res, err := json.Marshal(e)
	if err != nil {
		slog.Error("failed to marshal http exception", slog.Any("error", err))
		return
	}

	logStr := ""
	if e.Details != nil {
		logStr = fmt.Sprintf("Msg: %s; Status: %d; Error:\n%+v\nDetails:%v", e.PublicMessage, e.StatusCode, e.Error, e.Details)
	} else {
		logStr = fmt.Sprintf("Msg: %s; Status: %d; Error:\n%+v", e.PublicMessage, e.StatusCode, e.Error)
	}
	slog.Debug(logStr)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.StatusCode)
	w.Write(res)
}
