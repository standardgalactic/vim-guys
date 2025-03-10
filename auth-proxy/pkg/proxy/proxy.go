package proxy

import (
	"log/slog"
	"sync"

	"vim-guys.theprimeagen.tv/pkg/protocol"
)

type Interceptor interface {
	Id() int
}

type IRelay interface {
	AddInterceptor(i Interceptor)
	PushToGame(frame *protocol.ProtocolFrame) error
	PushToClient(frame *protocol.ProtocolFrame) error
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

func (r *Relay) PushToGame(frame *protocol.ProtocolFrame) error {
	slog.Info("to game", "frame", frame)
	return nil
}

func (r *Relay) PushToClient(frame *protocol.ProtocolFrame) error {
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

