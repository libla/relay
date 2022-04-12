package g

import (
	"unsafe"
)

func getg() unsafe.Pointer

func Get() unsafe.Pointer {
	return getg()
}
