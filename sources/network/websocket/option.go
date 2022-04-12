package websocket

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	defaultBufferSize  = 4096
	defaultPacketSize  = 65536
	defaultReadPacket  = 5
	defaultWritePacket = 100
)

type Option struct {
	enable         bool
	bufferSize     int
	maxPacketSize  int
	maxReadPacket  int
	maxWritePacket int
	keepAlive      time.Duration
}

func DefaultOption() Option {
	return Option{
		enable:         true,
		bufferSize:     defaultBufferSize,
		maxPacketSize:  defaultPacketSize,
		maxReadPacket:  defaultReadPacket,
		maxWritePacket: defaultWritePacket,
	}
}

func (this *Option) SetBufferSize(size int) *Option {
	if size > 0 {
		this.bufferSize = size
	}
	return this
}

func (this *Option) SetMaxPacketSize(size int) *Option {
	if size > 0 {
		this.maxPacketSize = size
	}
	return this
}

func (this *Option) SetMaxPacket(read, write int) *Option {
	if read > 0 {
		this.maxReadPacket = read
	}
	if write > 0 {
		this.maxWritePacket = write
	}
	return this
}

func (this *Option) SetKeepAlive(duration time.Duration) {
	this.keepAlive = duration
}

func (this *Option) apply(conn *websocket.Conn) {
	conn.SetReadLimit(int64(this.maxPacketSize))
}
