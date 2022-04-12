package websocket

import (
	"net"
	"net/http"
	"relay"
	"relay/codec"
	"relay/network"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

type Server interface {
	Start(address string) error
	Stop()
	State() network.State
}

type server[TInput, TOutput any] struct {
	loop       relay.Loop
	server     *http.Server
	option     Option
	exit       sync.WaitGroup
	state      int32
	upgrader   websocket.Upgrader
	encoder    func(codec.PipelineContext[Message], TOutput) error
	decoder    func(codec.PipelineContext[TInput], Message) error
	handle     network.ListenerHandle[TInput, TOutput]
	bufferpool relay.BufferPool
}

func Bind[TInput, TOutput any](option Option,
	encoder func(codec.PipelineContext[Message], TOutput) error,
	decoder func(codec.PipelineContext[TInput], Message) error,
	handle network.ListenerHandle[TInput, TOutput]) Server {
	if !option.enable {
		option = DefaultOption()
	}
	upgrader := websocket.Upgrader{
		ReadBufferSize:  option.bufferSize,
		WriteBufferSize: option.bufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	return &server[TInput, TOutput]{state: int32(network.Stopped), upgrader: upgrader, encoder: encoder, decoder: decoder, handle: handle, bufferpool: relay.NewBufferPool(option.bufferSize)}
}

func (this *server[TInput, TOutput]) Start(address string) error {
	loop := relay.InLoop()
	this.exit.Wait()
	if !atomic.CompareAndSwapInt32(&this.state, int32(network.Stopped), int32(network.Running)) {
		return nil
	}
	listener, err := net.Listen("tcp", address)
	if err != nil {
		atomic.CompareAndSwapInt32(&this.state, int32(network.Running), int32(network.Stopped))
		return errors.WithStack(err)
	}
	this.loop = loop
	this.server = &http.Server{Addr: address, Handler: handleFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrade := false
		connections := strings.Split(r.Header.Get("Connection"), ",")
		for _, connection := range connections {
			if strings.EqualFold(strings.TrimSpace(connection), "Upgrade") {
				upgrade = true
				break
			}
		}
		if !upgrade || !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
			w.WriteHeader(404)
			if flush, ok := w.(http.Flusher); ok {
				flush.Flush()
			}
			return
		}
		conn, err := this.upgrader.Upgrade(w, r, nil)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			if flush, ok := w.(http.Flusher); ok {
				flush.Flush()
			}
			return
		}
		this.option.apply(conn)
		session := &session[TInput, TOutput]{loop: this.loop, conn: conn, url: r.URL, header: r.Header, option: this.option, encoder: this.encoder, decoder: this.decoder, handle: this.handle, bufferpool: this.bufferpool}
		session.init()
		this.loop.Execute(relay.ExecFunc(func() error {
			return this.handle.OnAccept(session)
		}))
	})}
	this.exit.Add(1)
	go func() {
		defer this.exit.Done()
		this.server.Serve(listener)
	}()
	return nil
}

func (this *server[TInput, TOutput]) Stop() {
	if !atomic.CompareAndSwapInt32(&this.state, int32(network.Running), int32(network.Stopping)) {
		return
	}
	this.server.Close()
	this.exit.Wait()
}

func (this *server[TInput, TOutput]) State() network.State {
	return network.State(this.state)
}

type handleFunc func(http.ResponseWriter, *http.Request)

func (f handleFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f(w, r)
}
