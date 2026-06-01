package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

var ErrTokenNotFound = errors.New("extension not found")

func RedirectController(ctx context.Context, token string) (string, error) {
	url, err := Queries.GetURL(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%w: %w", ErrTokenNotFound, err)
		}
		return "", err
	}
	return url, nil
}

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	url, err := RedirectController(r.Context(), token)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		slog.Info("user redirected", "url", url)
		if err := Queries.AddTimestamp(r.Context(), token); err != nil {
			slog.Error("server failed to save timestamp", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, url, http.StatusFound)
	}
}
