package handlers

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
)

type NewURLRequest struct {
	URL string `json:"url"`
}

type NewURLResponse struct {
	URLToken string `json:"url_token"`
}

var ErrURLConflict = fmt.Errorf("url already exists")

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

func NewURLController(url string) (string, error) {
	// Generate unique token
	var token string
	for {
		generatedToken, err := generateToken()
		if err != nil {
			slog.Error("error in generating new token", "error", err)
			return "", err
		}
		if _, ok := db[generatedToken]; !ok {
			token = generatedToken
			break
		}
	}

	// Store in DB
	db[token] = url
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
	urlToken, err := NewURLController(payload.URL)
	if err != nil {
		if errors.Is(err, ErrURLConflict) {
			slog.Warn("user attempted to create redirect with already existing URL", "url", payload.URL)
			http.Error(w, "url already exists", http.StatusConflict)
		} else {
			slog.Error("an error occured when creating a new URL", "error", err)
			http.Error(w, "the server encountered an error", http.StatusInternalServerError)
		}
		return
	}

	// Send response with new token
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(NewURLResponse{URLToken: urlToken}); err != nil {
		slog.Error("server failed to encode JSON body", "error", err)
		http.Error(w, "the server encountered an error", http.StatusInternalServerError)
	} else {
		slog.Info("new url created", "url", payload.URL, "token", urlToken)
	}
}
