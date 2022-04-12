package network

import "errors"

type State int32

const (
	Stopped State = iota
	Running
	Stopping
)

var (
	Busy         = errors.New("network busy")
	Disconnected = errors.New("network disconnected")
)
