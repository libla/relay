package tcp

import (
	"io"
	"net"
	"relay"
	"relay/codec"
	"relay/network"
	"sync"
	"sync/atomic"
	"time"
)

type session[TInput, TOutput any] struct {
	loop        relay.Loop
	conn        *net.TCPConn
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
	encoder     func(codec.PipelineContext[relay.Buffer], TOutput) error
	decoder     func(codec.PipelineContext[TInput], relay.Buffer) error
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

func (this sessionOutput[TInput, TOutput]) Next(output relay.Buffer) error {
	for !output.Empty() {
		bytes, err := output.BeginRead()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		n, err := this.session.conn.Write(bytes)
		if err != nil {
			output.EndRead(0)
			if netErr, ok := err.(net.Error); ok && !netErr.Timeout() && netErr.Temporary() {
				time.Sleep(time.Millisecond)
				continue
			}
			return err
		}
		output.EndRead(n)
	}
	return nil
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
		this.conn.CloseRead()
		err := this.conn.SetReadDeadline(time.Now())
		if err != nil {
			return err
		}
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
					this.conn.CloseRead()
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
	buffer := this.bufferpool.New()
	for {
		bytes, err := buffer.BeginWrite()
		if err != nil {
			this.conn.SetWriteDeadline(time.Now())
			break
		}
		keepAlive := false
		if this.option.keepAlive > 0 {
			keepAlive = true
			this.conn.SetReadDeadline(time.Now().Add(this.option.keepAlive))
		}
		n, err := this.conn.Read(bytes)
		if keepAlive {
			this.conn.SetReadDeadline(time.Time{})
		}
		if err != nil {
			if netErr, ok := err.(net.Error); ok && !netErr.Timeout() && netErr.Temporary() {
				buffer.EndWrite(0)
				time.Sleep(time.Millisecond)
				continue
			}
			this.conn.SetWriteDeadline(time.Now())
			break
		}
		if !buffer.EndWrite(n) {
			this.conn.SetWriteDeadline(time.Now())
			break
		}
		err = this.decoder(this.readCtx, buffer)
		if err != nil {
			this.loop.Execute(relay.ExecFunc(func() error {
				this.handle.OnError(this, err)
				return nil
			}))
		}
	}
}
