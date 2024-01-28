package cmnerr

import (
	"log/slog"
	"net/http"

	"github.com/MykolaSainiuk/schatgo/src/common/httpexp"
)

func LogAndReply500(w http.ResponseWriter, err error) {
	slog.Error("Unknown error happened", slog.Any("error", err))
	Reply500(w, err)
}

func Reply500(w http.ResponseWriter, err error) {
	httpexp.FromText("It's a shame. We beg your pardon!", http.StatusInternalServerError).Reply(w)
}
