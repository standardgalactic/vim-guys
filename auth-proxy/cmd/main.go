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

func main() {
	godotenv.Load()

	ctx := config.NewAuthConfig(config.ProxyConfigParamsFromEnv())
	e := echo.New()
	e.GET("/socket", addWS(ctx))

	url := fmt.Sprintf("0.0.0.0:%d", config.Config.Port)
	if err := e.Start(url); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("echo server crashed", "error", err)
	}
}
