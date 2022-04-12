package main

import (
	"examples/benchmark"
	"fmt"
	"math/rand"
	"relay"
	"relay/network"
	"relay/network/tcp"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func main() {
	loop := relay.StartLoop()
	pipeline := benchmark.NullString{}
	loop.Execute(relay.ExecFunc(func() error {
		times := make(map[network.Session[string]]time.Time)
		var counts []time.Duration
		connect := tcp.Dial(tcp.DefaultOption(), pipeline.Encode, pipeline.Decode, network.NewSessionHandle(
			func(session network.Session[string], input string) error {
				last, ok := times[session]
				if ok {
					counts = append(counts, time.Since(last))
				}
				return nil
			},
			nil, nil))
		var timers []relay.Timer
		runes := make([]rune, 1024)
		for i := 0; i < 10000; i++ {
			for i := range runes {
				runes[i] = letters[rand.Intn(len(letters))]
			}
			text := string(runes)
			session, err := connect.Connect("127.0.0.1:8888")
			if err != nil {
				return err
			}
			session.Start()
			timer := relay.NewTimer(relay.TimeoutFunc(func(timer relay.Timer) error {
				err := session.Send(text)
				if err != nil {
					return err
				}
				times[session] = time.Now()
				return nil
			}))
			timers = append(timers, timer)
		}
		now := time.Now()
		for _, timer := range timers {
			timer.Start(now, relay.WithInterval(time.Millisecond*time.Duration(rand.Intn(250)+250)))
		}
		timer := relay.NewTimer(relay.TimeoutFunc(func(timer relay.Timer) error {
			sum := int(0)
			for _, count := range counts {
				sum += int(count)
			}
			fmt.Printf("average: %v, count: %d\n", time.Duration(sum/len(counts)), len(counts))
			counts = counts[0:0]
			return nil
		}))
		timer.StartNow(relay.WithInterval(time.Second * 1))
		return nil
	}))
	err := relay.Bootstrap(relay.EmptyConfig())
	if err != nil {
		panic(err)
	}
}
