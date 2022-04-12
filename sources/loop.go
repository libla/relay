package relay

import (
	"context"
	"fmt"
	"reflect"
	"relay/internal/g"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/pkg/errors"
)

type Executor interface {
	Execute() error
}

type Loop interface {
	context.Context
	Cancel()
	Execute(executor Executor) error
	Load(key any) (any, bool)
	Store(key, value any)
	Delete(key any)
}

type ExecFunc func() error

func (this ExecFunc) Execute() error {
	return this()
}

func InLoop() Loop {
	loop := currentloop()
	if loop == nil {
		panic(errors.New("can only be called in a Loop"))
	}
	return loop
}

func IsInLoop() Loop {
	return currentloop()
}

func Poll[T any](ch <-chan T) (result T, err error) {
	this := InLoop().(*loop)
	defer func() {
		if r := recover(); r != nil {
			result, ok := r.(error)
			if ok {
				if _, ok := result.(stackTracer); ok {
					err = result
				} else {
					err = errors.WithStack(result)
				}
			} else {
				err = errors.Errorf("%v", r)
			}
		}
	}()
	co := this.current
	this.current = nil
	this.self.resume(nil)
	result = <-ch
	this.list.pushfront(co.executor)
	co.yield()
	return
}

func Await(action any, params ...any) ([]any, error) {
	this := InLoop().(*loop)
	co := this.current
	this.current = nil
	this.self.resume(nil)
	results, err := protect(action, params...)
	this.list.pushfront(co.executor)
	co.yield()
	return results, err
}

func protect(action any, params ...any) (results []any, err error) {
	defer func() {
		if r := recover(); r != nil {
			result, ok := r.(error)
			if ok {
				err = result
			} else {
				err = errors.Errorf("%v", r)
			}
		}
	}()
	actionType := reflect.TypeOf(action)
	actionValue := reflect.ValueOf(action)
	numIn := actionType.NumIn()
	numOut := actionType.NumOut()
	returnsLength := numOut
	if numOut != 0 {
		if actionType.Out(numOut-1) == errorType {
			returnsLength--
		}
	}
	var out []reflect.Value
	if numIn == 0 {
		out = actionValue.Call(nil)
	} else {
		argsLength := len(params)
		argumentIn := numIn
		if actionType.IsVariadic() {
			argumentIn--
		}

		if argsLength < argumentIn {
			panic(errors.New("with too few input arguments"))
		}

		in := make([]reflect.Value, numIn)
		for i := 0; i < argumentIn; i++ {
			if params[i] == nil {
				in[i] = reflect.Zero(actionType.In(i))
			} else {
				in[i] = reflect.ValueOf(params[i])
			}
		}

		if actionType.IsVariadic() {
			m := argsLength - argumentIn
			slice := reflect.MakeSlice(actionType.In(numIn-1), m, m)
			in[numIn-1] = slice
			for i := 0; i < m; i++ {
				x := params[argumentIn+i]
				if x != nil {
					slice.Index(i).Set(reflect.ValueOf(x))
				}
			}
			out = actionValue.CallSlice(in)
		} else {
			out = actionValue.Call(in)
		}
	}

	if out != nil {
		if returnsLength != 0 {
			results = make([]interface{}, returnsLength)
			for i := 0; i < returnsLength; i++ {
				results[i] = out[i].Interface()
			}
		}
		if returnsLength != len(out) {
			result := out[returnsLength].Interface()
			if result != nil {
				err = result.(error)
			}
		}
	}
	return
}

func StartLoop(errors ...func(error) error) Loop {
	context, cancel := context.WithCancel(Application)
	result := &loop{Context: context, cancel: cancel, errors: errors, self: coroutine{signal: make(chan Executor)}}
	result.list.init()
	loops.Add(1)
	go func() {
		defer result.list.pushback(nil)
		<-context.Done()
	}()
	go func() {
		defer loops.Done()
		for {
			executor := result.list.pop()
			if executor == nil {
				break
			}
			if result.Err() != nil {
				break
			}
			co := result.getfree()
			result.current = co
			co.resume(executor)
			result.self.yield()
		}
		for {
			var co *coroutine
			result.lock.Lock()
			if result.freelist != nil {
				co = result.freelist
				result.freelist = co.next
			}
			result.lock.Unlock()
			if co == nil {
				break
			}
			co.resume(nil)
			result.self.yield()
		}
	}()
	return result
}

type loop struct {
	context.Context
	cancel   context.CancelFunc
	errors   []func(error) error
	values   sync.Map
	list     tasklist
	lock     sync.Mutex
	count    int32
	self     coroutine
	current  *coroutine
	freelist *coroutine
}

func currentloop() *loop {
	pointer := g.Get()
	value, ok := goloops.Load(pointer)
	if !ok {
		return nil
	}
	loop, ok := value.(*loop)
	if !ok {
		return nil
	}
	if loop.current == nil || loop.current.pointer != pointer {
		return nil
	}
	return loop
}

func (this *loop) Cancel() {
	this.cancel()
}

func (this *loop) Execute(executor Executor) error {
	if executor == nil {
		return nil
	}
	err := this.Err()
	if err != nil {
		return err
	}
	this.list.pushback(executor)
	return nil
}

func (this *loop) Load(key any) (any, bool) {
	return this.values.Load(key)
}

func (this *loop) Store(key, value any) {
	this.values.Store(key, value)
}

func (this *loop) Delete(key any) {
	this.values.Delete(key)
}

func (this *loop) DebugPrint() {
	this.list.guard.Lock()
	defer this.list.guard.Unlock()
	fmt.Println(fmt.Sprintf("pending: %d", this.list.len))
}

func (this *loop) getfree() *coroutine {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.freelist != nil {
		co := this.freelist
		this.freelist = co.next
		return co
	}
	co := &coroutine{signal: make(chan Executor)}
	co.executor = ExecFunc(func() error {
		this.current = co
		co.resume(nil)
		this.self.yield()
		return nil
	})
	atomic.AddInt32(&this.count, 1)
	wait := make(chan struct{})
	go func() {
		co.pointer = g.Get()
		goloops.Store(co.pointer, this)
		wait <- Void
		defer func() {
			atomic.AddInt32(&this.count, -1)
			goloops.Delete(co.pointer)
			this.self.resume(nil)
			if result := recover(); result != nil {
				if wrap, ok := result.(errorWrap); ok {
					errloop <- wrap.err
				} else {
					if err, ok := result.(error); ok {
						for _, error := range this.errors {
							err = error(err)
							if err == nil {
								break
							}
						}
						if err != nil {
							errloop <- err
						}
					} else {
						panic(result)
					}
				}
			}
		}()
		for {
			executor := co.yield()
			if executor == nil {
				break
			}
			err := executor.Execute()
			if err != nil {
				for _, error := range this.errors {
					err = error(err)
					if err == nil {
						break
					}
				}
				if err != nil {
					panic(errorWrap{err})
				}
			}
			this.putfree(co)
			this.self.resume(nil)
		}
	}()
	<-wait
	close(wait)
	return co
}

func (this *loop) putfree(co *coroutine) {
	this.lock.Lock()
	defer this.lock.Unlock()
	co.next = this.freelist
	this.freelist = co
}

type errorWrap struct {
	err error
}

type coroutine struct {
	pointer  unsafe.Pointer
	signal   chan Executor
	executor Executor
	next     *coroutine
}

func (this *coroutine) yield() Executor {
	return <-this.signal
}

func (this *coroutine) resume(executor Executor) {
	this.signal <- executor
}

type tasklist struct {
	root   tasknode
	len    int
	guard  sync.Mutex
	signal *sync.Cond
}

type tasknode struct {
	list     *tasklist
	prev     *tasknode
	next     *tasknode
	executor Executor
}

var taskfreelist = sync.Pool{
	New: func() any {
		return &tasknode{}
	},
}

func (this *tasklist) init() {
	this.root.next = &this.root
	this.root.prev = &this.root
	this.len = 0
	this.signal = sync.NewCond(&this.guard)
}

func (this *tasklist) pushfront(executor Executor) {
	node := taskfreelist.Get().(*tasknode)
	node.executor = executor
	this.guard.Lock()
	this.insert(node, &this.root)
}

func (this *tasklist) pushback(executor Executor) {
	node := taskfreelist.Get().(*tasknode)
	node.executor = executor
	this.guard.Lock()
	this.insert(node, this.root.prev)
}

func (this *tasklist) insert(node, at *tasknode) {
	node.prev = at
	node.next = at.next
	at.next = node
	node.next.prev = node
	node.list = this
	this.len++
	this.guard.Unlock()
	this.signal.Signal()
}

func (this *tasklist) pop() Executor {
	this.guard.Lock()
	defer this.guard.Unlock()
	for this.len == 0 {
		this.signal.Wait()
	}
	this.len--
	node := this.root.next
	node.prev.next = node.next
	node.next.prev = node.prev
	executor := node.executor
	node.recycle()
	return executor
}

func (this *tasknode) recycle() {
	this.next = nil
	this.prev = nil
	this.list = nil
	this.executor = nil
	taskfreelist.Put(this)
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()
var goloops sync.Map
