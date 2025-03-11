package server

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"vim-guys.theprimeagen.tv/pkg/config"
)

// Upgrader configures the WebSocket connection
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Allow all origins
}

func addWS(ctx config.ProxyContext) func(c echo.Context) error {
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


