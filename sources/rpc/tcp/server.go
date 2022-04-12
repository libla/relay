package tcp

import (
	"relay/codec"
	"relay/rpc"
)

type Server interface {
	rpc.Server
}

func Bind[Request, Response any](server codec.Response[Request, Response, []byte, []byte]) Server {
	return nil
}
