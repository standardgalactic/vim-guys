package ws

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"vim-guys.theprimeagen.tv/pkg/config"
	"vim-guys.theprimeagen.tv/pkg/data"
	"vim-guys.theprimeagen.tv/pkg/protocol"
	"vim-guys.theprimeagen.tv/pkg/proxy"
)


type WSFactory struct {
	websocketId atomic.Int64
	context *config.ProxyContext
	logger *slog.Logger
}

type WS struct {
	conn   *websocket.Conn
	closed bool
	websocketId int
	mutex  sync.Mutex
	context *config.ProxyContext
	proxy proxy.IProxy
	logger *slog.Logger
}

func NewWSProducer(c *config.ProxyContext) *WSFactory {
	return &WSFactory{
		websocketId: atomic.Int64{},
		context: c,
		logger: slog.Default().With("area", "ws-factory"),
	}
}

func (p *WSFactory) NewWS(conn *websocket.Conn) *WS {
	id := int(p.websocketId.Add(1))
	return &WS{
		conn:   conn,
		closed: false,
		websocketId: id,
		mutex:  sync.Mutex{},
		context: p.context,
		logger: slog.Default().With("area", fmt.Sprintf("ws-%d", id)),
	}
}

func (w *WS) Id() int {
	return w.websocketId
}

func (w *WS) Name() string {
	return "websocket"
}

func (w *WS) Start(p proxy.IProxy) error {
	if !w.context.HasDatabase() {
		return fmt.Errorf("unable to create a websocket connection without a database.  unable to perform authentication.")
	}
	w.proxy = p
	err := w.authenticate()
	if err != nil {
		w.logger.Debug("unable to authenticate websocket client", "id", w.Id())
	}

	// listen for messages and pass them to the game
	for {
		frame, err := w.next()
		if err != nil {
			slog.Debug("websocket errored", "error", err)
			w.Close()
			break
		}
		slog.Debug("received frame", "frame", frame, "id", w.Id())

		// TODO filter for things the proxy can understand (stat requests, game quitting, etc etc)
		// TODO pass the rest that make sense to the game
		p.PushToGame(frame, w)
	}
	return nil
}

func (w *WS) ToClient(frame *protocol.ProtocolFrame) error {
	// TODO lets see if i can keep this
	// I may have to do some magic and probably rename "Original" into frame data
	return w.conn.WriteMessage(websocket.BinaryMessage, frame.Frame())
}

func (w *WS) next() (*protocol.ProtocolFrame, error) {
	for {
		t, data, err := w.conn.ReadMessage()
		w.logger.Info("msg received", "type", t, "data length", len(data), "err", err)
		if err != nil {
			return nil, err
		}

		if t != websocket.BinaryMessage {
			continue
		}

		frame, err := protocol.FromData(data, w.websocketId)
		w.logger.Info("msg parsed", "frame", frame, "error", err)
		return frame, err
	}
}

func (w *WS) Close() error {
	if w.proxy != nil {
		w.proxy.RemoveInterceptor(w)
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.closed = true
	w.conn.Close()
	return nil
}

func (w *WS) authenticate() error {
	ctx, cancel := context.WithTimeout(context.Background(), w.context.WS.AuthenticationTimeout)
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
		err := w.context.DB.Get(&mapping, query, token)

		if err != nil {
			w.logger.Error("Failed to select user_mapping", "error", err)
			return err
		}


	}
	return nil
}
