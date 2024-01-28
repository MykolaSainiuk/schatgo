package middleware

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/docgen"
)

const docPath string = "./readme/file.md"

func Readme(r chi.Router) http.HandlerFunc {
	doca := docgen.MarkdownRoutesDoc(r, docgen.MarkdownOpts{
		ProjectPath: "github.com/MykolaSainiuk/schatgo",
		Intro:       "Welcome to the schatgo API documentation",
	})

	writeDownDoc(docPath, &doca)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/markdown; charset=UTF-8")
		http.ServeFile(w, r, docPath)
	})
}

func writeDownDoc(fpath string, content *string) {
	f, err := os.OpenFile(fpath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		slog.Error("cannot open file", slog.Any("error", err.Error()))
		return
	}
	defer f.Close()

	f.WriteString(*content)
	f.Sync()
}
