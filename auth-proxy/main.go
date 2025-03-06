package main

import (
	"io"
	"log/slog"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		res, err := http.Get("http://vimguys.dev.vim-guys.sst:8080")
		if err != nil {
			slog.Error("unable to get vimguys", "error", err)
			w.Write([]byte("SADGE DAYS"))
			return
		}
		body, err := io.ReadAll(res.Body)
		if err != nil {
			slog.Error("unable to decode", "error", err)
			w.Write([]byte("SADGE DAYS, unable to read body"))
			return
		}
		w.Write([]byte("Hello, World!" + string(body)))
	})

	http.ListenAndServe(":8080", nil)
}

