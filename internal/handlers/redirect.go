package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
)

var db = map[string]string{
	"g": "https://www.google.com",
}

var ErrExtensionNotFound = fmt.Errorf("extension not found")

func RedirectController(ext string) (string, error) {
	if url, ok := db[ext]; !ok {
		slog.Warn(ErrExtensionNotFound.Error(), "ext", ext)
		return "", ErrExtensionNotFound
	} else {
		return url, nil
	}
}

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	url, err := RedirectController(r.PathValue("ext"))

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		slog.Info("user redirected", "url", url)
		http.Redirect(w, r, url, http.StatusFound)
	}
}
