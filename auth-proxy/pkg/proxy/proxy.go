package proxy

import (
	"log/slog"
	"slices"
	"sync"

	"vim-guys.theprimeagen.tv/pkg/config"
	"vim-guys.theprimeagen.tv/pkg/protocol"
)

type Interceptor interface {
	Id() int
	Start(IProxy) error
	Name() string
	Close() error
}

type IProxy interface {
	AddInterceptor(i Interceptor)
	RemoveInterceptor(i Interceptor)
	PushToGame(*protocol.ProtocolFrame, Interceptor) error
	PushToClient(*protocol.ProtocolFrame) error
	Context() *config.ProxyContext
}

type Proxy struct {
	// TODO interceptors organized by game_id and by player_id
	// There is going to be specific messages and broadcast messages
	mutex sync.Mutex
	interceptors []Interceptor
	context *config.ProxyContext
}

func (r *Proxy) NewProxy(ctx *config.ProxyContext) *Proxy {
	return &Proxy{
		mutex: sync.Mutex{},
		interceptors: []Interceptor{},
		context: ctx,
	}
}

func (r *Proxy) Context() *config.ProxyContext {
	return r.context
}

func (r *Proxy) PushToGame(frame *protocol.ProtocolFrame, i Interceptor) error {
	slog.Info("to game", "frame", frame, "from", i.Id())
	return nil
}

func (r *Proxy) PushToClient(frame *protocol.ProtocolFrame) error {
	slog.Info("to client", "frame", frame)
	return nil
}

func (r *Proxy) RemoveInterceptor(i Interceptor) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	found := slices.DeleteFunc(r.interceptors, func(i Interceptor) bool {
		return i.Id() == i.Id()
	})

	for _, v := range found {
		v.Close()
	}
}

func (r *Proxy) AddInterceptor(i Interceptor) {
	go func() {
		err := i.Start(r)
		if err != nil {
			slog.Error("Interceptor failed to start", "name", i.Name(), "id", i.Id())
			r.RemoveInterceptor(i)
		}
	}()
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.interceptors = append(r.interceptors, i)
}


func (r *Proxy) Start() error {
	// TODO I should be reading from messages from game and messages from clients
	return nil
}

