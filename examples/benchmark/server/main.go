package main

import (
	"examples/benchmark"
	"relay"
	"relay/network"
	"relay/network/tcp"
	"time"
)

type print interface {
	DebugPrint()
}

func main() {
	loop := relay.StartLoop()
	print := loop.(print)
	pipeline := benchmark.NullString{}
	loop.Execute(relay.ExecFunc(func() error {
		timer := relay.NewTimer(relay.TimeoutFunc(func(timer relay.Timer) error {
			print.DebugPrint()
			return nil
		}))
		timer.StartNow(relay.WithInterval(time.Second * 2))
		server := tcp.Bind(tcp.DefaultOption(), pipeline.Encode, pipeline.Decode, network.NewListenerHandle(
			func(session network.Session[string], input string) error {
				return session.Send(input)
			},
			nil, nil, nil))
		return server.Start(":8888")
	}))
	err := relay.Bootstrap(relay.EmptyConfig())
	if err != nil {
		panic(err)
	}
}
