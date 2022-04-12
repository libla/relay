package tcp

import (
	"net"
	"time"
)

const (
	defaultBufferSize  = 4096
	defaultReadPacket  = 5
	defaultWritePacket = 100
)

type Option struct {
	enable         bool
	noDelay        bool
	bufferSize     int
	maxReadPacket  int
	maxWritePacket int
	keepAlive      time.Duration
}

func DefaultOption() Option {
	return Option{
		enable:         true,
		bufferSize:     defaultBufferSize,
		maxReadPacket:  defaultReadPacket,
		maxWritePacket: defaultWritePacket,
	}
}

func (this *Option) SetNoDelay(value bool) *Option {
	this.noDelay = value
	return this
}

func (this *Option) SetBufferSize(size int) *Option {
	if size > 0 {
		this.bufferSize = size
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

func (this *Option) apply(conn *net.TCPConn) {
	conn.SetNoDelay(this.noDelay)
}
