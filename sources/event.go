package relay

import (
	"container/list"
	"strings"
)

type Event struct {
	listeners *list.List
}

type EventArg[T any] struct {
	listeners *list.List
}

type EventHandle interface {
	Handle() error
}

type EventArgHandle[T any] interface {
	Handle(T) error
}

func (this *Event) Listen(listener func() error) func() {
	if this.listeners == nil {
		this.listeners = list.New()
	}
	element := this.listeners.PushBack(listener)
	return func() {
		this.listeners.Remove(element)
	}
}

func (this Event) Emit() error {
	if this.listeners == nil {
		return nil
	}
	var errors []error
	element := this.listeners.Front()
	for element != nil {
		listener := element.Value.(func() error)
		err := listener()
		if err != nil {
			errors = append(errors, err)
		}
		element = element.Next()
	}
	if len(errors) == 0 {
		return nil
	}
	return &eventerror{list: errors}
}

func (this *EventArg[T]) Listen(listener func(T) error) func() {
	if this.listeners == nil {
		this.listeners = list.New()
	}
	element := this.listeners.PushBack(listener)
	return func() {
		this.listeners.Remove(element)
	}
}

func (this EventArg[T]) Emit(value T) error {
	if this.listeners == nil {
		return nil
	}
	var errors []error
	element := this.listeners.Front()
	for element != nil {
		listener := element.Value.(func(T) error)
		err := listener(value)
		if err != nil {
			errors = append(errors, err)
		}
		element = element.Next()
	}
	if len(errors) == 0 {
		return nil
	}
	return &eventerror{list: errors}
}

type eventerror struct {
	list   []error
	errmsg *string
}

func (this *eventerror) Error() string {
	if this.errmsg == nil {
		list := make([]string, len(this.list))
		for i := 0; i < len(this.list); i++ {
			list[i] = this.list[i].Error()
		}
		result := strings.Join(list, "\n")
		this.errmsg = &result
	}
	return *this.errmsg
}
