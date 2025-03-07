package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"vim-guys.theprimeagen.tv/pkg/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/tursodatabase/libsql-client-go/libsql" // Register the libsql driver
)

// UserMapping represents the structure of the user_mapping table
type UserMapping struct {
    UserID     string `db:"USERID"`
    UUID string `db:"UUID"`
}

func (u *UserMapping) String() string {
	return fmt.Sprintf("UserId: %s -- UUID: %s", u.UserID, u.UUID)
}

type Interceptor interface {
	Id() int
}

type IRelay interface {
	AddInterceptor(i Interceptor)
	PushToGame(frame *ProtocolFrame) error
	PushToClient(frame *ProtocolFrame) error
}

type Relay struct {
	mutex sync.Mutex
	interceptors []Interceptor
}

func (r *Relay) NewRelay() *Relay {
	return &Relay{
		mutex: sync.Mutex{},
		interceptors: []Interceptor{},
	}
}

func (r *Relay) PushToGame(frame *ProtocolFrame) error {
	slog.Info("to game", "frame", frame)
	return nil
}

func (r *Relay) PushToClient(frame *ProtocolFrame) error {
	slog.Info("to client", "frame", frame)
	return nil
}

func (r *Relay) AddInterceptor(i Interceptor) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.interceptors = append(r.interceptors, i)
	return nil
}


func (r *Relay) Start() error {
	// i think there is something here i should be doing
	return nil
}

type ProtocolType int
const VERSION = 1
const HEADER_LENGTH = 14
const (
	Authenticate ProtocolType = 1
	Authenticated ProtocolType = 2
)

var VersionMismatch = fmt.Errorf("protocol version mismatch")
var UnknownType = fmt.Errorf("unknown type")
var MalformedFrame = fmt.Errorf("malformed frame")
var LengthMismatch = fmt.Errorf("length mismatch")
var playerId atomic.Int64

func isType(t ProtocolType) bool {
	// TODO better than this??
	return t >= Authenticate && t <= Authenticated
}

type ProtocolFrame struct {
	Type ProtocolType `json:"type"`
	Len int `json:"len"`
	Data []byte `json:"data"`
	PlayerId int `json:"playerId"`
	GameId int `json:"gameId"` // TODO: wtf do i do here...
	Original []byte
}

func NewProtocolFrame(t ProtocolType, data []byte, playerId int) *ProtocolFrame {
	return &ProtocolFrame{
		Type: t,
		Len: len(data),
		Data: data,
		PlayerId: playerId,
	}
}

func FromData(data []byte, playerId int) (*ProtocolFrame, error) {
	if len(data) < HEADER_LENGTH {
		return nil, MalformedFrame
	}
	original := data[0:]

	version := binary.BigEndian.Uint16(data)
	if version != VERSION {
		return nil, VersionMismatch
	}

	data = data[2:]
	t := ProtocolType(binary.BigEndian.Uint16(data))
	if !isType(t) {
		return nil, UnknownType
	}

	data = data[2:]
	length := int(binary.BigEndian.Uint16(data))

	data = data[2:] // move forward by length

	data = data[8:] // erases playerid + gameid
	slog.Info("length parsed", "length", length, "data remaining", len(data))

	if len(data) != length {
		return nil, LengthMismatch
	}

	return &ProtocolFrame{
		Type: t,
		Len: length,
		Data: data,
		PlayerId: playerId,
		GameId: 0,
		Original: original,
	}, nil
}

func (f *ProtocolFrame) Frame() []byte {
	// TODO still probably bad idea...
	if f.Original != nil {
		return f.Original
	}

	length := HEADER_LENGTH + f.Len
	data := make([]byte, length, length)

	writer := data[:HEADER_LENGTH]
	binary.BigEndian.PutUint16(writer, VERSION)

	writer = writer[2:]
	binary.BigEndian.PutUint16(writer, uint16(f.Type))

	// TODO write assert lib
	writer = writer[2:]
	binary.BigEndian.PutUint16(writer, uint16(f.Len))

	writer = writer[2:]
	binary.BigEndian.PutUint32(writer, uint32(f.PlayerId))

	copy(data[HEADER_LENGTH:], f.Data)
	return data
}

// Upgrader configures the WebSocket connection
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Allow all origins
}

type WS struct {
	conn   *websocket.Conn
	closed bool
	playerId int
	mutex  sync.Mutex
}

func NewWS(conn *websocket.Conn) *WS {
	return &WS{
		conn:   conn,
		closed: false,
		playerId: int(playerId.Add(1)),
		mutex:  sync.Mutex{},
	}
}

func (w *WS) Id() int {
	return w.playerId
}

func (w *WS) ToClient(frame *ProtocolFrame) error {
	// TODO lets see if i can keep this
	// I may have to do some magic and probably rename "Original" into frame data
	return w.conn.WriteMessage(websocket.BinaryMessage, frame.Frame())
}

func (w *WS) next() (*ProtocolFrame, error) {
	for {
		t, data, err := w.conn.ReadMessage()
		slog.Info("msg received", "type", t, "data length", len(data), "err", err)
		if err != nil {
			return nil, err
		}

		if t != websocket.BinaryMessage {
			continue
		}

		frame, err := FromData(data, w.playerId)
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
	ctx, cancel := context.WithTimeout(outer, config.Config.AuthenticationTimeout)
	next := make(chan *ProtocolFrame, 1)
	go func() {
		data, err := w.next()
		if err == nil {
			next <- data
		}
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("socket didn't respond in time")
	case msg := <-next:
		cancel()
		if msg.Type != Authenticate {
			return fmt.Errorf("expected authentication packet but received: %d", msg.Type)
		}
		token := string(msg.Data)
		slog.Info("token received", "token", token)
		query := "SELECT userId, uuid FROM user_mapping WHERE uuid = ?"
		var mapping UserMapping
		err := db.Get(&mapping, query, token)
		if err != nil {
			slog.Error("Failed to select user_mapping", "error", err)
			return err
		}

		slog.Info("user mapping result", "mapping", mapping.String())

	}
	return nil
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

	slog.Info("DB", "url", dbURL)

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
    slog.Info("Successfully connected to Turso database!")

	playerId.Store(0)
	e := echo.New()
	e.GET("/socket", addWS(db))

	url := fmt.Sprintf("0.0.0.0:%d", config.Config.Port)
	if err := e.Start(url); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("echo server crashed", "error", err)
	}
}
