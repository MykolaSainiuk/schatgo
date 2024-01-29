package httpexp

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
)

// InvalidResponseDto
//
//	@Description	common 422 respons
type HttpExp struct {
	Error      error `json:"-"`
	StatusCode int   `json:"-"`

	Message string   `json:"message,omitempty"`
	Details []string `json:"details"`
}

func FromError(err error, code int, details ...string) *HttpExp {
	return &HttpExp{
		Error:      err,
		StatusCode: code,
		Message:    err.Error(),
		Details:    details,
	}
}

func FromText(message string, code int, details ...string) *HttpExp {
	return &HttpExp{
		Error:      nil,
		StatusCode: code,
		Message:    message,
		Details:    details,
	}
}

func (e *HttpExp) SetMessage(message string) *HttpExp {
	e.Message = message
	return e
}

func (e *HttpExp) Reply(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.StatusCode)
	res, err := json.Marshal(e)
	if err != nil {
		slog.Error("failed to marshal http exception", slog.Any("error", err))
		return
	}
	if e.Details != nil {
		slog.Debug(e.Message,
			slog.Int("status", int(e.StatusCode)),
			slog.Any("error", e.Error),
			slog.String("msg", strings.Join(e.Details, "; ")),
		)
	} else {
		slog.Debug(e.Message,
			slog.Int("status", int(e.StatusCode)),
			slog.Any("error", e.Error),
		)
	}

	w.Write(res)
}
