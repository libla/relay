package http

import (
	"relay/codec"
	"relay/rpc"
)

type Client interface {
	rpc.Client
}

func Connect[Request, Response any](server codec.Request[Request, Response, []byte, []byte], url string) Client {
	return nil
}
