package relay

import (
	"reflect"
	"sync"
)

func Lock[T comparable](first T, rest ...T) func() {
	return Locks(append([]T{first}, rest...))
}

func Locks[T comparable](params []T) func() {
	this := InLoop().(*loop)
	typeof := reflect.TypeOf((*T)(nil)).Elem()
	result := &lock[T]{values: params}
	result.wait.Add(1)
	if !result.tryLock(true) {
		value, ok := typelists.Load(typeof)
		if !ok {
			value, _ = typelists.LoadOrStore(typeof, &typelist[T]{
				locks: make(map[T]struct{}),
				waits: make(map[T][]*lock[T]),
			})
		}
		addLock(value.(*typelist[T]), result, params)
	}
	co := this.current
	this.current = nil
	this.self.resume(nil)
	result.wait.Wait()
	this.list.pushfront(co.executor)
	co.yield()
	return func() {
		value, ok := typelists.Load(typeof)
		if ok {
			list := value.(*typelist[T])
			list.guard.Lock()
			defer list.guard.Unlock()
			for _, param := range params {
				delete(list.locks, param)
			}
			for _, param := range params {
				locks, ok := list.waits[param]
				if ok {
					for _, lock := range locks {
						if lock.tryLock(false) {
							break
						}
					}
				}
			}
		}
	}
}

func addLock[T comparable](list *typelist[T], result *lock[T], values []T) {
	list.guard.Lock()
	defer list.guard.Unlock()
	for _, value := range values {
		locks, ok := list.waits[value]
		if ok {
			locks = append(locks, result)
		} else {
			locks = []*lock[T]{result}
		}
		list.waits[value] = locks
	}
}

var typelists sync.Map

type typelist[T comparable] struct {
	guard sync.Mutex
	locks map[T]struct{}
	waits map[T][]*lock[T]
}

type lock[T comparable] struct {
	values []T
	wait   sync.WaitGroup
}

func (this *lock[T]) tryLock(guard bool) bool {
	typeof := reflect.TypeOf((*T)(nil)).Elem()
	value, ok := typelists.Load(typeof)
	if !ok {
		value, _ = typelists.LoadOrStore(typeof, &typelist[T]{
			locks: make(map[T]struct{}),
			waits: make(map[T][]*lock[T]),
		})
	}
	list := value.(*typelist[T])
	if guard {
		list.guard.Lock()
		defer list.guard.Unlock()
	}
	if len(list.locks) != 0 {
		for _, value := range this.values {
			_, exists := list.locks[value]
			if exists {
				return false
			}
		}
	}
	for _, value := range this.values {
		list.locks[value] = Void
		locks, ok := list.waits[value]
		if ok {
			for i := 0; i < len(locks); i++ {
				if locks[i] == this {
					locks = append(locks[:i], locks[i+1:]...)
					list.waits[value] = locks
					break
				}
			}
		}
	}
	this.wait.Done()
	return true
}
