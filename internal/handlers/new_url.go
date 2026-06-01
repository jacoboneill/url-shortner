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
	"net/url"

	"github.com/jacoboneill/url-shortner/internal/db"
)

type NewURLRequest struct {
	URL   *string `json:"url"`
	Title *string `json:"title"`
}

type NewURLResponse struct {
	URLToken string `json:"url_token"`
}

var (
	ErrUniqueTokenGeneration = errors.New("failed to generate unique token")
	ErrInvalidURL            = errors.New("URL missing from request body")
	ErrInvalidURLScheme      = errors.New("invalid url scheme")
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

func NewURLController(ctx context.Context, url string, title *string) (string, error) {
	// Store in DB
	token, err := generateUniqueToken(ctx)
	if err != nil {
		return "", fmt.Errorf("%w, %w", ErrUniqueTokenGeneration, err)
	}

	createURLParams := db.CreateURLParams{
		Token: token,
		Url:   url,
	}
	if title != nil {
		createURLParams.Title = *title
	}

	if err := Queries.CreateURL(ctx, createURLParams); err != nil {
		return "", err
	}

	return token, nil
}

func validateURL(URL *string) (string, error) {
	if URL == nil {
		return "", ErrInvalidURL
	}

	u, err := url.ParseRequestURI(*URL)
	if err != nil {
		return *URL, err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", ErrInvalidURLScheme
	}
	return u.String(), nil
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

	// Validate URL
	URL, err := validateURL(payload.URL)
	if err != nil {
		if errors.Is(err, ErrInvalidURL) {
			slog.Warn("user attempted to request with no URL")
			http.Error(w, "invalid JSON", http.StatusBadRequest)
		} else {
			slog.Warn("user attempted to request with bad URL", "url", URL, "error", err)
			http.Error(w, "invalid URL", http.StatusBadRequest)
		}
		return
	}

	// Get new token
	urlToken, err := NewURLController(r.Context(), URL, payload.Title)
	if err != nil {
		if errors.Is(err, ErrUniqueTokenGeneration) {
			slog.Error("server failed to generate a unique token", "error", err)
		} else {
			slog.Error("an unknown error occured when creating a new URL", "error", err)
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Send response with new token
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(NewURLResponse{URLToken: urlToken}); err != nil {
		slog.Error("server failed to encode JSON body", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	} else {
		slog.Info("new url created", "url", URL, "token", urlToken)
	}
}
