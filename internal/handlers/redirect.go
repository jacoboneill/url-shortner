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
	url, err := RedirectController(r.Context(), r.PathValue("token"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		slog.Info("user redirected", "url", url)
		http.Redirect(w, r, url, http.StatusFound)
	}
}
