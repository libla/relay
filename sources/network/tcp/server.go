package tcp

import (
	"net"
	"relay"
	"relay/codec"
	"relay/network"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

type Server interface {
	Start(address string) error
	Stop()
	State() network.State
}

type server[TInput, TOutput any] struct {
	loop       relay.Loop
	listener   net.Listener
	option     Option
	exit       sync.WaitGroup
	state      int32
	encoder    func(codec.PipelineContext[relay.Buffer], TOutput) error
	decoder    func(codec.PipelineContext[TInput], relay.Buffer) error
	handle     network.ListenerHandle[TInput, TOutput]
	bufferpool relay.BufferPool
}

func Bind[TInput, TOutput any](option Option,
	encoder func(codec.PipelineContext[relay.Buffer], TOutput) error,
	decoder func(codec.PipelineContext[TInput], relay.Buffer) error,
	handle network.ListenerHandle[TInput, TOutput]) Server {
	if !option.enable {
		option = DefaultOption()
	}
	return &server[TInput, TOutput]{state: int32(network.Stopped), encoder: encoder, decoder: decoder, handle: handle, bufferpool: relay.NewBufferPool(option.bufferSize)}
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
	this.listener = listener
	this.exit.Add(1)
	go this.accept()
	return nil
}

func (this *server[TInput, TOutput]) Stop() {
	if !atomic.CompareAndSwapInt32(&this.state, int32(network.Running), int32(network.Stopping)) {
		return
	}
	this.listener.Close()
	this.exit.Wait()
}

func (this *server[TInput, TOutput]) State() network.State {
	return network.State(this.state)
}

func (this *server[TInput, TOutput]) accept() {
	defer func() {
		if atomic.CompareAndSwapInt32(&this.state, int32(network.Stopping), int32(network.Stopped)) {
			this.exit.Done()
		}
		atomic.CompareAndSwapInt32(&this.state, int32(network.Running), int32(network.Stopped))
	}()
	for {
		conn, err := this.listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				time.Sleep(time.Millisecond)
				continue
			}
			/*
				if opErr, ok := err.(*net.OpError); ok {
					switch e := opErr.Err.(type) {
					case *os.SyscallError:
						if errno, ok := e.Err.(syscall.Errno); ok {
							switch errno {
							case syscall.ETIMEDOUT, syscall.ECONNABORTED, syscall.ECONNRESET, syscall.EINTR,
								syscall.EMFILE, syscall.ENFILE, syscall.EAGAIN, syscall.EWOULDBLOCK, syscall.EBUSY:
								time.Sleep(time.Millisecond)
								continue
							}
						}
					}
				}
			*/
			break
		}
		connTcp := conn.(*net.TCPConn)
		this.option.apply(connTcp)
		session := &session[TInput, TOutput]{loop: this.loop, conn: connTcp, option: this.option, encoder: this.encoder, decoder: this.decoder, handle: this.handle, bufferpool: this.bufferpool}
		session.init()
		this.loop.Execute(relay.ExecFunc(func() error {
			return this.handle.OnAccept(session)
		}))
	}
}
