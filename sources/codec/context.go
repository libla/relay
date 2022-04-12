package codec

import "relay"

type Context interface {
	Close() error
	Load(key any) (value any, ok bool)
	Store(key, value any)
	Delete(key any)
	Alloc() relay.Buffer
}
