package main

import (
	"bufio"
	"examples/echo"
	"os"
	"relay"
	"relay/network"
	"relay/network/tcp"
	"strings"
)

func main() {
	loop := relay.StartLoop()
	pipeline := echo.NullString{}
	loop.Execute(relay.ExecFunc(func() error {
		connect := tcp.Dial(tcp.DefaultOption(), pipeline.Encode, pipeline.Decode, network.NewSessionHandle(
			func(session network.Session[string], input string) error {
				println("-> " + input)
				return nil
			},
			nil, nil))
		session, err := connect.Connect("127.0.0.1:8888")
		if err != nil {
			return err
		}
		session.Start()
		go func() {
			for {
				text, err := bufio.NewReader(os.Stdin).ReadString('\n')
				if err != nil {
					continue
				}
				text = strings.TrimSpace(text)
				session.Send(text)
			}
		}()
		return nil
	}))
	err := relay.Bootstrap(relay.EmptyConfig())
	if err != nil {
		panic(err)
	}
}
