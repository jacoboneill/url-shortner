package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log/slog"
	"net/http"

	db "github.com/jacoboneill/url-shortner/internal/db"
	"github.com/jacoboneill/url-shortner/internal/handlers"
	_ "modernc.org/sqlite"
)

func initDatabase() (*db.Queries, error) {
	const dbFP = "/data/main.db"

	conn, err := sql.Open("sqlite", dbFP)
	if err != nil {
		return nil, err
	}

	_, err = conn.ExecContext(context.Background(), db.DDL)
	if err != nil {
		return nil, err
	}

	return db.New(conn), nil
}

func initMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{token}", handlers.RedirectHandler)
	mux.HandleFunc("POST /", handlers.NewURLHandler)

	return mux
}

func main() {
	mux := initMux()
	queries, err := initDatabase()
	if err != nil {
		slog.Error("failed to initialise database", "error", err)
		return
	}
	handlers.Queries = queries

	const port = 8000
	slog.Info("starting server", "port", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		slog.Error("server failed to start", "err", err)
	}
}
