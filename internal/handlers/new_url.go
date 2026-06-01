package handlers

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"

	"github.com/jacoboneill/url-shortner/internal/db"
)

type NewURLRequest struct {
	URL string `json:"url"`
}

type NewURLResponse struct {
	URLToken string `json:"url_token"`
}

var (
	ErrURLConflict           = errors.New("url already exists")
	ErrUniqueTokenGeneration = errors.New("failed to generate unique token")
)

func generateToken() (string, error) {
	const (
		charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		length  = 6
	)

	token := make([]byte, length)
	for i := range token {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		token[i] = charset[n.Int64()]
	}
	return string(token), nil
}

func generateUniqueToken(ctx context.Context) (string, error) {
	for {
		token, err := generateToken()
		if err != nil {
			return "", err
		}
		existsInt, err := Queries.TokenExists(ctx, token)
		if err != nil {
			return "", err
		}
		if exists := existsInt != 0; !exists {
			return token, nil
		}
	}
}

func NewURLController(ctx context.Context, url string) (string, error) {
	// Store in DB
	token, err := generateUniqueToken(ctx)
	if err != nil {
		return "", fmt.Errorf("%w, %w", ErrUniqueTokenGeneration, err)
	}

	Queries.CreateURL(ctx, db.CreateURLParams{
		Token: token,
		Url:   url,
	})
	return token, nil
}

func NewURLHandler(w http.ResponseWriter, r *http.Request) {
	// Check content type
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		slog.Warn("user attempted to request with invalid content type", "Content-Type", contentType)
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	// Decode request body's JSON
	var payload NewURLRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		slog.Warn("server failed to decode JSON body", "body", r.Body, "error", err)
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	// Get new token
	urlToken, err := NewURLController(r.Context(), payload.URL)
	if err != nil {
		if errors.Is(err, ErrURLConflict) {
			slog.Warn("user attempted to create redirect with already existing URL", "url", payload.URL)
			http.Error(w, "bad request", http.StatusConflict)
		} else {
			if errors.Is(err, ErrUniqueTokenGeneration) {
				slog.Error("server failed to generate a unique token", "error", err)
			} else {
				slog.Error("an unknown error occured when creating a new URL", "error", err)
			}
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Send response with new token
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(NewURLResponse{URLToken: urlToken}); err != nil {
		slog.Error("server failed to encode JSON body", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	} else {
		slog.Info("new url created", "url", payload.URL, "token", urlToken)
	}
}
