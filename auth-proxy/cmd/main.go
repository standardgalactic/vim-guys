package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"vim-guys.theprimeagen.tv/pkg/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/tursodatabase/libsql-client-go/libsql" // Register the libsql driver
)

// Upgrader configures the WebSocket connection
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Allow all origins
}

func addWS(db *sqlx.DB) func(c echo.Context) error {
	return func(c echo.Context) error {
		// Upgrade the HTTP connection to a WebSocket connection
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			slog.Error("WebSocket upgrade error:", "error", err)
			return err
		}

		ws := NewWS(conn)
		err = ws.authenticate(context.Background(), db)
		if err != nil {
			slog.Error("Websocket authenticate failed", "error", err)
			err = ws.ToClient(NewProtocolFrame(Authenticated, []byte{0}, ws.playerId))
			ws.Close()
			return err
		}

		err = ws.ToClient(NewProtocolFrame(Authenticated, []byte{1}, ws.playerId))
		slog.Error("Websocket authenticate successed", "send error", err)
		ws.Close()
		return nil
	}
}

func main() {
	godotenv.Load()

	dbURL := os.Getenv("TURSO_DATABASE_URL")
    dbToken := os.Getenv("TURSO_AUTH_TOKEN")

	// ASSERT LIBRARY YOU DUMMY
	connStr := fmt.Sprintf("libsql://%s?authToken=%s", dbURL, dbToken)
	db, err := sqlx.Connect("libsql", connStr)
    if err != nil {
        slog.Error("Failed to connect to Turso database", "error", err)
		return
    }
    defer db.Close()
	// Test the connection
    if err := db.Ping(); err != nil {
        slog.Error("Failed to ping database", "error", err)
    }
    slog.Warn("Successfully connected to Turso database!")

	playerId.Store(0)
	e := echo.New()
	e.GET("/socket", addWS(db))

	url := fmt.Sprintf("0.0.0.0:%d", config.Config.Port)
	if err := e.Start(url); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("echo server crashed", "error", err)
	}
}
