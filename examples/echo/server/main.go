package main

import (
	"examples/echo"
	"os"
	"relay"
	"relay/network"
	"relay/network/tcp"
)

func main() {
	var prefix string
	if len(os.Args) > 1 {
		prefix = os.Args[1] + ": "
	}
	loop := relay.StartLoop()
	pipeline := echo.NullString{}
	loop.Execute(relay.ExecFunc(func() error {
		server := tcp.Bind(tcp.DefaultOption(), pipeline.Encode, pipeline.Decode, network.NewListenerHandle(
			func(session network.Session[string], input string) error {
				return session.Send(prefix + input)
			},
			nil, nil, nil))
		return server.Start(":8888")
	}))
	err := relay.Bootstrap(relay.EmptyConfig())
	if err != nil {
		panic(err)
	}
}
