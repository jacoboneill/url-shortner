package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jacoboneill/url-shortner/internal/handlers"
)

func main() {
	const port = 8000

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{ext}", handlers.RedirectHandler)

	slog.Info("starting server", "port", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		slog.Error("server failed to start", "err", err)
	}
}
