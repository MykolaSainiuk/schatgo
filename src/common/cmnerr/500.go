package cmnerr

import (
	"net/http"

	"github.com/MykolaSainiuk/schatgo/src/common/httpexp"
)

func Reply500(w http.ResponseWriter, err error) {
	httpexp.From(err, "It's a shame. We beg your pardon!", http.StatusInternalServerError).Reply(w)
}
