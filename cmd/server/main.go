package main

import (
	"fmt"
	"log/slog"
	"net/http"
)

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world!"))
}

func main() {
	const port = 8000

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", helloWorld)

	slog.Info("starting server", "port", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		slog.Error("server failed to start", "err", err)
	}
}
