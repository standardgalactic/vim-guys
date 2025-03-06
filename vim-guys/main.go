package main

import (
	"log/slog"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Request received")
		w.Write([]byte("from vim guys"))
	})

	http.ListenAndServe(":8080", nil)
}

