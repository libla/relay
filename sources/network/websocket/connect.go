package websocket

import (
	"net/http"
	"net/url"
	"relay"
	"relay/codec"
	"relay/network"

	"github.com/gorilla/websocket"
)

type Connector[TValue any] interface {
	Connect(address string, header http.Header) (network.Session[TValue], error)
}

type connector[TInput, TOutput any] struct {
	option  Option
	encoder func(codec.PipelineContext[Message], TOutput) error
	decoder func(codec.PipelineContext[TInput], Message) error
	handle  network.SessionHandle[TInput, TOutput]
}

type dialresult struct {
	conn *websocket.Conn
	err  error
}

func Dial[TInput, TOutput any](option Option,
	encoder func(codec.PipelineContext[Message], TOutput) error,
	decoder func(codec.PipelineContext[TInput], Message) error,
	handle network.SessionHandle[TInput, TOutput]) Connector[TOutput] {
	if !option.enable {
		option = DefaultOption()
	}
	return &connector[TInput, TOutput]{option, encoder, decoder, handle}
}

func (this *connector[TInput, TOutput]) Connect(address string, header http.Header) (network.Session[TOutput], error) {
	url, err := url.ParseRequestURI(address)
	if err != nil {
		return nil, err
	}
	loop := relay.InLoop()
	ch := make(chan dialresult)
	go func(address string) {
		conn, _, err := websocket.DefaultDialer.DialContext(loop, address, header)
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
	conn := result.conn
	this.option.apply(conn)
	session := &session[TInput, TOutput]{loop: loop, conn: conn, url: url, header: header, option: this.option, encoder: this.encoder, decoder: this.decoder, handle: this.handle, bufferpool: relay.NewBufferPool(this.option.bufferSize)}
	session.init()
	return session, nil
}
