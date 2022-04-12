package echo

import (
	"io"
	"relay"
	"relay/codec"
)

var nullstringpoll = relay.NewBufferPool(256)

var nullstringkey struct{}

type NullString struct {
}

func (this NullString) Encode(context codec.PipelineContext[relay.Buffer], output string) error {
	buffer := nullstringpoll.New()
	_, err := buffer.Write([]byte(output))
	if err != nil {
		return err
	}
	buffer.WriteByte(0)
	return context.Next(buffer)
}

func (this NullString) Decode(context codec.PipelineContext[string], input relay.Buffer) error {
	var buffer relay.Buffer
	value, ok := context.Load(&nullstringkey)
	if ok {
		buffer, ok = value.(relay.Buffer)
		if !ok {
			buffer = nullstringpoll.New()
			context.Store(&nullstringkey, buffer)
		}
	} else {
		buffer = nullstringpoll.New()
		context.Store(&nullstringkey, buffer)
	}
	for {
		c, err := input.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if c == 0 {
			var result []byte
			for {
				bytes, err := buffer.BeginRead()
				if err != nil {
					context.Delete(&nullstringkey)
					return err
				}
				l := len(bytes)
				if l == 0 {
					buffer.EndRead(0)
					break
				}
				result = append(result, bytes...)
				buffer.EndRead(l)
			}
			context.Next(string(result))
		} else {
			buffer.WriteByte(c)
		}
	}
	return nil
}
