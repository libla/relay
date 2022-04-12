package tcp

import (
	"relay/codec"
	"relay/msg"
)

type Client interface {
	msg.Message
}

func Connect[T any](server codec.Pipeline[[]byte, T], address string) Client {
	return nil
}
