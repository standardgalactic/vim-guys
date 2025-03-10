package ws

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"vim-guys.theprimeagen.tv/pkg/config"
	"vim-guys.theprimeagen.tv/pkg/data"
	"vim-guys.theprimeagen.tv/pkg/protocol"
)


type WSFactory struct {
	websocketId atomic.Int64
	context *config.ProxyContext
}

type WS struct {
	conn   *websocket.Conn
	closed bool
	websocketId int
	mutex  sync.Mutex
	context *config.ProxyContext
}

func NewWSProducer(c *config.ProxyContext) *WSFactory {
	return &WSFactory{
		websocketId: atomic.Int64{},
		context: c,
	}
}

func (p *WSFactory) NewWS(conn *websocket.Conn) *WS {
	return &WS{
		conn:   conn,
		closed: false,
		websocketId: int(p.websocketId.Add(1)),
		mutex:  sync.Mutex{},
		context: p.context,
	}
}

func (w *WS) Id() int {
	return w.websocketId
}

func (w *WS) ToClient(frame *protocol.ProtocolFrame) error {
	// TODO lets see if i can keep this
	// I may have to do some magic and probably rename "Original" into frame data
	return w.conn.WriteMessage(websocket.BinaryMessage, frame.Frame())
}

func (w *WS) next() (*protocol.ProtocolFrame, error) {
	for {
		t, data, err := w.conn.ReadMessage()
		slog.Info("msg received", "type", t, "data length", len(data), "err", err)
		if err != nil {
			return nil, err
		}

		if t != websocket.BinaryMessage {
			continue
		}

		frame, err := protocol.FromData(data, w.websocketId)
		slog.Info("msg parsed", "frame", frame, "error", err)
		return frame, err
	}
}

func (w *WS) Close() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.closed = true
	w.conn.Close()
}

func (w *WS) authenticate(outer context.Context, db *sqlx.DB) error {
	ctx, cancel := context.WithTimeout(outer, w.context.WS.AuthenticationTimeout)
	next := make(chan *protocol.ProtocolFrame, 1)
	go func() {
		data, err := w.next()
		if err == nil {
			next <- data
		}
	}()

	select {
	case <-ctx.Done():
		cancel()
		return errors.New("socket didn't respond in time")
	case msg := <-next:
		cancel()
		if msg.Type != protocol.Authenticate {
			return fmt.Errorf("expected authentication packet but received: %d", msg.Type)
		}
		token := string(msg.Data)

		query := "SELECT userId, uuid FROM user_mapping WHERE uuid = ?"
		var mapping data.UserMapping
		err := db.Get(&mapping, query, token)

		if err != nil {
			slog.Error("Failed to select user_mapping", "error", err)
			return err
		}


	}
	return nil
}
