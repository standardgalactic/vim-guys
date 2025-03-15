package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"vim-guys.theprimeagen.tv/auth-proxy/pkg/config"
	"vim-guys.theprimeagen.tv/auth-proxy/pkg/proxy"
	"vim-guys.theprimeagen.tv/auth-proxy/pkg/ws"
)

// Upgrader configures the WebSocket connection
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Allow all origins
}

type ProxyServer struct { }

func NewProxyServer() *ProxyServer {
	return &ProxyServer{ }
}

func (p *ProxyServer) Id() int {
	return config.PROXY_SERVER_ID
}

func (prox *ProxyServer) Start(p proxy.IProxy) error {
	factory :=  ws.NewWSProducer(p.Context())
	e := echo.New()

	e.GET("/health", func(c echo.Context) error {
		return nil
	})

	e.GET("/socket", func(c echo.Context) error {
		// Upgrade the HTTP connection to a WebSocket connection
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			slog.Error("WebSocket upgrade error:", "error", err)
			return err
		}

		ws := factory.NewWS(conn)
		p.AddInterceptor(ws)
		return nil
	})

	url := fmt.Sprintf("0.0.0.0:%d", p.Context().Port)
	if err := e.Start(url); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("echo server crashed", "error", err)
	}
}

func (p *ProxyServer) Name() string { return "ProxyServer" }
func (p *ProxyServer) Close() error {
	return nil
}


