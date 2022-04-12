package websocket

import (
	"relay/codec"
	"relay/msg"
)

type Server interface {
	msg.Message
}

func Bind[T any](server codec.Pipeline[[]byte, T]) Server {
	return nil
}
