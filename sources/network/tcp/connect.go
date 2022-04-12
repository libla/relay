package tcp

import (
	"net"
	"relay"
	"relay/codec"
	"relay/network"
)

type Connector[TValue any] interface {
	Connect(address string) (network.Session[TValue], error)
}

type connector[TInput, TOutput any] struct {
	option  Option
	encoder func(codec.PipelineContext[relay.Buffer], TOutput) error
	decoder func(codec.PipelineContext[TInput], relay.Buffer) error
	handle  network.SessionHandle[TInput, TOutput]
}

type dialresult struct {
	conn net.Conn
	err  error
}

func Dial[TInput, TOutput any](option Option,
	encoder func(codec.PipelineContext[relay.Buffer], TOutput) error,
	decoder func(codec.PipelineContext[TInput], relay.Buffer) error,
	handle network.SessionHandle[TInput, TOutput]) Connector[TOutput] {
	if !option.enable {
		option = DefaultOption()
	}
	return &connector[TInput, TOutput]{option, encoder, decoder, handle}
}

func (this *connector[TInput, TOutput]) Connect(address string) (network.Session[TOutput], error) {
	loop := relay.InLoop()
	ch := make(chan dialresult)
	go func(address string) {
		var dial net.Dialer
		conn, err := dial.DialContext(loop, "tcp", address)
		ch <- dialresult{conn, err}
	}(address)
	result, err := relay.Poll(ch)
	close(ch)
	if err != nil {
		return nil, err
	}
	if result.err != nil {
		return nil, result.err
	}
	conn := result.conn.(*net.TCPConn)
	this.option.apply(conn)
	session := &session[TInput, TOutput]{loop: loop, conn: conn, option: this.option, encoder: this.encoder, decoder: this.decoder, handle: this.handle, bufferpool: relay.NewBufferPool(this.option.bufferSize)}
	session.init()
	return session, nil
}
