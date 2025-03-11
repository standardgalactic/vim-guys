package protocol

import (
	"encoding/binary"
	"fmt"
	"log/slog"
)

var VersionMismatch = fmt.Errorf("protocol version mismatch")
var UnknownType = fmt.Errorf("unknown type")
var MalformedFrame = fmt.Errorf("malformed frame")
var LengthMismatch = fmt.Errorf("length mismatch")

type ProtocolType int
const VERSION = 1
const HEADER_LENGTH = 14
const (
	Authenticate ProtocolType = 1
	Authenticated ProtocolType = 2
)

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

func Auth(auth bool, playerId int) *ProtocolFrame {
	b := byte(0)
	if auth {
		b = 1
	}
	return NewProtocolFrame(Authenticate, []byte{b}, playerId)
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


