package websocket

import (
	"net"
	"net/http"
	"net/url"
	"relay"
	"relay/codec"
	"relay/network"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type MsgType int

const (
	TextMessage   = MsgType(websocket.TextMessage)
	BinaryMessage = MsgType(websocket.BinaryMessage)
	PingMessage   = MsgType(websocket.PingMessage)
	PongMessage   = MsgType(websocket.PongMessage)
)

type Message struct {
	Type MsgType
	Body []byte
}

type Session[TValue any] interface {
	network.Session[TValue]
	URL() *url.URL
	Header() http.Header
}

type session[TInput, TOutput any] struct {
	loop        relay.Loop
	conn        *websocket.Conn
	url         *url.URL
	header      http.Header
	option      Option
	runflag     int32
	flagMessage int32
	execMessage relay.Executor
	exit        sync.WaitGroup
	values      sync.Map
	readList    chan TInput
	writeList   chan TOutput
	readCtx     sessionInput[TInput, TOutput]
	writeCtx    sessionOutput[TInput, TOutput]
	encoder     func(codec.PipelineContext[Message], TOutput) error
	decoder     func(codec.PipelineContext[TInput], Message) error
	handle      network.SessionHandle[TInput, TOutput]
	bufferpool  relay.BufferPool
}

type sessionInput[TInput, TOutput any] struct {
	session *session[TInput, TOutput]
}

type sessionOutput[TInput, TOutput any] struct {
	session *session[TInput, TOutput]
}

func (this sessionInput[TInput, TOutput]) Next(input TInput) error {
	this.session.readList <- input
	this.session.message()
	return nil
}

func (this sessionInput[TInput, TOutput]) Close() error {
	return this.session.Close()
}

func (this sessionInput[TInput, TOutput]) Load(key any) (value any, ok bool) {
	return this.session.values.Load(key)
}

func (this sessionInput[TInput, TOutput]) Store(key, value any) {
	this.session.values.Store(key, value)
}

func (this sessionInput[TInput, TOutput]) Delete(key any) {
	this.session.values.Delete(key)
}

func (this sessionInput[TInput, TOutput]) Alloc() relay.Buffer {
	return this.session.bufferpool.New()
}

func (this sessionOutput[TInput, TOutput]) Next(output Message) error {
	return this.session.conn.WriteMessage(int(output.Type), output.Body)
}

func (this sessionOutput[TInput, TOutput]) Close() error {
	return this.session.Close()
}

func (this sessionOutput[TInput, TOutput]) Load(key any) (value any, ok bool) {
	return this.session.values.Load(key)
}

func (this sessionOutput[TInput, TOutput]) Store(key, value any) {
	this.session.values.Store(key, value)
}

func (this sessionOutput[TInput, TOutput]) Delete(key any) {
	this.session.values.Delete(key)
}

func (this sessionOutput[TInput, TOutput]) Alloc() relay.Buffer {
	return this.session.bufferpool.New()
}

func (this *session[TInput, TOutput]) init() {
	if !this.option.enable {
		this.option = DefaultOption()
	}
	this.readList = make(chan TInput, this.option.maxReadPacket)
	this.writeList = make(chan TOutput, this.option.maxWritePacket)
	this.readCtx = sessionInput[TInput, TOutput]{session: this}
	this.writeCtx = sessionOutput[TInput, TOutput]{session: this}
	this.execMessage = relay.ExecFunc(func() error {
		atomic.StoreInt32(&this.flagMessage, 0)
		for {
			select {
			case input := <-this.readList:
				err := this.handle.OnMessage(this, input)
				if err != nil {
					this.message()
					return err
				}
			default:
				return nil
			}
		}
	})
}

func (this *session[TInput, TOutput]) Start() error {
	if atomic.CompareAndSwapInt32(&this.runflag, 0, 1) {
		this.exit.Add(2)
		go this.cleanup()
		go this.read()
		go this.write()
	}
	return nil
}

func (this *session[TInput, TOutput]) Close() error {
	if atomic.CompareAndSwapInt32(&this.runflag, 1, 2) {
		this.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second*2))
		return this.conn.Close()
	}
	return nil
}

func (this *session[TInput, TOutput]) Send(output TOutput) error {
	if atomic.LoadInt32(&this.runflag) != 1 {
		return network.Disconnected
	}
	select {
	case this.writeList <- output:
		return nil
	default:
		return network.Busy
	}
}

func (this *session[TInput, TOutput]) Connected() bool {
	return atomic.LoadInt32(&this.runflag) != 2
}

func (this *session[TInput, TOutput]) Started() bool {
	return atomic.LoadInt32(&this.runflag) == 1
}

func (this *session[TInput, TOutput]) URL() *url.URL {
	return this.url
}

func (this *session[TInput, TOutput]) Header() http.Header {
	return this.header
}

func (this *session[TInput, TOutput]) message() {
	if atomic.CompareAndSwapInt32(&this.flagMessage, 0, 1) {
		this.loop.Execute(this.execMessage)
	}
}

func (this *session[TInput, TOutput]) cleanup() {
	this.exit.Wait()
	atomic.StoreInt32(&this.runflag, 2)
	this.loop.Execute(relay.ExecFunc(func() error {
		err := this.handle.OnClose(this)
		go this.conn.Close()
		return err
	}))
}

func (this *session[TInput, TOutput]) write() {
	defer this.exit.Done()
	for {
		output := <-this.writeList
		err := this.encoder(this.writeCtx, output)
		if err != nil {
			if netErr, ok := err.(net.Error); ok {
				if netErr.Timeout() {
					this.conn.SetReadLimit(0)
					this.conn.SetReadDeadline(time.Now())
					break
				} else {
					this.loop.Execute(relay.ExecFunc(func() error {
						this.handle.OnError(this, err)
						return nil
					}))
				}
			} else {
				this.loop.Execute(relay.ExecFunc(func() error {
					this.handle.OnError(this, err)
					return nil
				}))
			}
		}
	}
}

func (this *session[TInput, TOutput]) read() {
	defer this.exit.Done()
	for {
		keepAlive := false
		if this.option.keepAlive > 0 {
			keepAlive = true
			this.conn.SetReadDeadline(time.Now().Add(this.option.keepAlive))
		}
		msgtype, bytes, err := this.conn.ReadMessage()
		if keepAlive {
			this.conn.SetReadDeadline(time.Time{})
		}
		if err != nil {
			this.conn.SetWriteDeadline(time.Now())
			break
		}
		err = this.decoder(this.readCtx, Message{MsgType(msgtype), bytes})
		if err != nil {
			this.loop.Execute(relay.ExecFunc(func() error {
				this.handle.OnError(this, err)
				return nil
			}))
		}
	}
}
