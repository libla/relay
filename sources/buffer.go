package relay

import (
	gerrors "errors"
	"io"
	"sync"

	"github.com/pkg/errors"
)

var ErrReading = gerrors.New("reading")
var ErrWriting = gerrors.New("writing")

type Buffer interface {
	Empty() bool
	Len() int
	Reset()
	io.Reader
	io.Writer
	io.ByteReader
	io.ByteWriter
	BeginRead() ([]byte, error)
	EndRead(read int) bool
	BeginWrite() ([]byte, error)
	EndWrite(wrote int) bool
}

type BufferPool interface {
	New() Buffer
}

func NewBufferPool(size int) BufferPool {
	if size <= 0 {
		panic(errors.New("pool size incorrect"))
	}
	pool := &bufferpool{}
	pool.pool.New = func() any {
		return &buffernode{bytes: make([]byte, size)}
	}
	return pool
}

type bufferpool struct {
	pool sync.Pool
}

func (this *bufferpool) New() Buffer {
	return &buffer{pool: this}
}

type buffer struct {
	pool    *bufferpool
	read    *buffernode
	write   *buffernode
	reading bool
	writing bool
}

type buffernode struct {
	bytes []byte
	read  int
	write int
	next  *buffernode
}

func (this *buffer) Empty() bool {
	if this.read == nil {
		return true
	}
	if this.read != this.write {
		return false
	}
	return this.read.read == this.read.write
}

func (this *buffer) Len() int {
	sum := 0
	cursor := this.read
	for cursor != nil {
		if cursor == this.write {
			sum += this.write.write - cursor.read
			break
		}
		sum += cursor.write - cursor.read
		cursor = cursor.next
	}
	return sum
}

func (this *buffer) Reset() {
	cursor := this.read
	for cursor != nil {
		next := cursor.next
		cursor.next = nil
		cursor.read = 0
		cursor.write = 0
		this.pool.pool.Put(cursor)
		cursor = next
	}
	this.read = nil
	this.write = nil
	this.reading = false
	this.writing = false
}

func (this *buffer) Read(bytes []byte) (int, error) {
	if this.reading {
		return 0, ErrReading
	}
	max := len(bytes)
	if max == 0 {
		return 0, nil
	}
	sum := 0
	if this.read != nil {
		for sum < max {
			cursor := this.read
			result := copy(bytes[sum:], cursor.bytes[cursor.read:cursor.write])
			cursor.read += result
			sum += result
			if cursor.read != cursor.write {
				continue
			}
			next := cursor.next
			if this.writing && next == nil {
				break
			}
			this.read = next
			cursor.next = nil
			cursor.read = 0
			cursor.write = 0
			this.pool.pool.Put(cursor)
			if this.read == nil {
				this.Reset()
				break
			}
		}
	}
	if sum == 0 {
		return 0, io.EOF
	}
	return sum, nil
}

func (this *buffer) ReadByte() (byte, error) {
	if this.reading {
		return 0, ErrReading
	}
	cursor := this.read
	if cursor == nil {
		return 0, io.EOF
	}
	if cursor.read == cursor.write {
		return 0, io.EOF
	}
	result := cursor.bytes[cursor.read]
	cursor.read++
	if cursor.read == cursor.write {
		next := cursor.next
		if !this.writing || next != nil {
			this.read = next
			cursor.next = nil
			cursor.read = 0
			cursor.write = 0
			this.pool.pool.Put(cursor)
			if this.read == nil {
				this.Reset()
			}
		}
	}
	return result, nil
}

func (this *buffer) Write(bytes []byte) (int, error) {
	if this.writing {
		return 0, ErrWriting
	}
	max := len(bytes)
	sum := 0
	for sum < max {
		cursor := this.write
		if cursor == nil {
			cursor = this.pool.pool.Get().(*buffernode)
			this.write = cursor
			this.read = cursor
		}
		if cursor.write == len(cursor.bytes) {
			cursor = this.pool.pool.Get().(*buffernode)
			this.write.next = cursor
			this.write = cursor
		}
		result := copy(cursor.bytes[cursor.write:], bytes[sum:])
		cursor.write += result
		sum += result
	}
	return sum, nil
}

func (this *buffer) WriteByte(c byte) error {
	if this.writing {
		return ErrWriting
	}
	cursor := this.write
	if cursor == nil {
		cursor = this.pool.pool.Get().(*buffernode)
		this.write = cursor
		this.read = cursor
	}
	if cursor.write == len(cursor.bytes) {
		cursor = this.pool.pool.Get().(*buffernode)
		this.write.next = cursor
		this.write = cursor
	}
	cursor.bytes[cursor.write] = c
	cursor.write++
	return nil
}

func (this *buffer) BeginRead() ([]byte, error) {
	if this.reading {
		return nil, ErrReading
	}
	this.reading = true
	if this.read == nil {
		return nil, nil
	}
	return this.read.bytes[this.read.read:this.read.write], nil
}

func (this *buffer) EndRead(read int) bool {
	if !this.reading {
		return false
	}
	this.reading = false
	if read == 0 {
		return true
	}
	cursor := this.read
	if cursor == nil {
		return false
	}
	if cursor.read+read > cursor.write {
		return false
	}
	cursor.read += read
	if cursor.read == cursor.write {
		next := cursor.next
		if !this.writing || next != nil {
			this.read = next
			cursor.next = nil
			cursor.read = 0
			cursor.write = 0
			this.pool.pool.Put(cursor)
			if this.read == nil {
				this.Reset()
			}
		}
	}
	return true
}

func (this *buffer) BeginWrite() ([]byte, error) {
	if this.writing {
		return nil, ErrWriting
	}
	this.writing = true
	cursor := this.write
	if cursor == nil {
		cursor = this.pool.pool.Get().(*buffernode)
		this.write = cursor
		this.read = cursor
	}
	if cursor.write == len(cursor.bytes) {
		cursor = this.pool.pool.Get().(*buffernode)
		this.write.next = cursor
		this.write = cursor
	}
	return cursor.bytes[cursor.write:], nil
}

func (this *buffer) EndWrite(wrote int) bool {
	if !this.writing {
		return false
	}
	this.writing = false
	if this.write.write+wrote > len(this.write.bytes) {
		return false
	}
	this.write.write += wrote
	return true
}
