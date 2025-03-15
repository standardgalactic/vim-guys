package intutils

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/tursodatabase/libsql-client-go/libsql"
	"vim-guys.theprimeagen.tv/auth-proxy/pkg/config"
	"vim-guys.theprimeagen.tv/auth-proxy/pkg/protocol"
	"vim-guys.theprimeagen.tv/auth-proxy/pkg/proxy"
	"vim-guys.theprimeagen.tv/auth-proxy/pkg/server"
)

func CreateDB(t *testing.T, f string) *sqlx.DB {
	connStr := fmt.Sprintf("file://%s", f)
	db, err := sqlx.Connect("libsql", connStr)
	require.NoError(t, err, "unable to make connection to database", "error", err)

	return db
}
func ReadNextBinary(conn *websocket.Conn, dur time.Duration) (*protocol.ProtocolFrame, error) {
	out := make(chan []byte, 1)
	errOut := make(chan error, 1)
	go func() {
		data, err := conn.ReadBinary()
		if err == nil {
			out <- data
		} else {
			errOut <- err
		}
	}()

	select {
	case <-time.NewTimer(dur).C:
		return nil, fmt.Errorf("time limit exceeded")
	case err := <- errOut:
		return nil, err
	case d := <- out:
		return protocol.FromData(d, 0)
	}
}

func CreateWS(t *testing.T, port uint, token string) (protocol.ProtocolFrame, *websocket.Conn) {
	conn, err := websocket.Dial(fmt.Sprintf("ws://localhost:%d", port))
	require.NoError(t, "unable to initialize websocket client connection", "error", err)

	auth := protocol.NewClientProtocolFrame(protocol.Authenticate, []byte(token))
	err = conn.Write(auth.Frame());
	require.NoError(t, "unable to send authentication packet", "error", err)

	p, err := ReadNextBinary(conn, time.Microsecond * 100)
	require.NoError(t, "error waiting for auth message back", "error", err)

	return p, conn
}

func LaunchServer(t *testing.T, c config.ProxyContext, maxAttempts int) (*proxy.Proxy, *server.ProxyServer) {
	p := proxy.NewProxy(c)
	s := server.NewProxyServer()
	p.AddInterceptor(s)

	for range maxAttempts {
		time.Sleep(time.Millisecond)
		_, err := http.Get(fmt.Sprintf("http://localhost:%d/health", c.Port))
		if err == nil {
			return p, s
		}
	}

	require.Never(t, "unable to launch server after waiting for maxAttempts", "maxAttempts", maxAttempts)
	return nil, nil
}
