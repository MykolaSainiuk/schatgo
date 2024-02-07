package httpexp

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
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

var (
	isProdOnce sync.Once
	isProd     = true
)

func From(err error, msg string, code int, details ...string) *HttpExp {
	isProdOnce.Do(func() {
		isProd = os.Getenv("NODE_ENV") == "production"
	})
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

	if !isProd {
		logStr := ""
		if e.Details == nil {
			logStr = fmt.Sprintf("Msg: %s; Status: %d; Error:\n%+v", e.PublicMessage, e.StatusCode, e.Error)
		} else {
			logStr = fmt.Sprintf("Msg: %s; Status: %d; Error:\n%+v\nDetails:%+v", e.PublicMessage, e.StatusCode, e.Error, e.Details)
		}
		slog.Debug(logStr)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.StatusCode)
	w.Write(res)
}
