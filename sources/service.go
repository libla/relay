package relay

import (
	"container/list"
	"errors"
)

type Service struct {
	implements *list.List
	last       *list.Element
}

type ServiceArg[Arg any] struct {
	implements *list.List
	last       *list.Element
}

type ServiceResult[Result any] struct {
	implements *list.List
	last       *list.Element
}

type ServiceArgResult[Arg, Result any] struct {
	implements *list.List
	last       *list.Element
}

func (this *Service) Register(implement func() error) func() {
	if this.implements == nil {
		this.implements = list.New()
	}
	element := this.implements.PushBack(implement)
	return func() {
		this.implements.Remove(element)
	}
}

func (this *Service) Call() error {
	if this.implements == nil {
		return NotImplement
	}
	start := this.last
	if start == nil || start.Prev() == nil {
		start = this.implements.Front()
	}
	err := NotImplement
	element := start
	for {
		if element == nil {
			element = this.implements.Front()
		}
		if element == start {
			break
		}
		implement := element.Value.(func() error)
		err = implement()
		if err == nil {
			this.last = element
			return nil
		}
		element = element.Next()
	}
	return err
}

func (this *ServiceArg[Arg]) Register(implement func(Arg) error) func() {
	if this.implements == nil {
		this.implements = list.New()
	}
	element := this.implements.PushBack(implement)
	return func() {
		this.implements.Remove(element)
	}
}

func (this *ServiceArg[Arg]) Call(arg Arg) error {
	if this.implements == nil {
		return NotImplement
	}
	start := this.last
	if start == nil || start.Prev() == nil {
		start = this.implements.Front()
	}
	err := NotImplement
	element := start
	for {
		if element == nil {
			element = this.implements.Front()
		}
		if element == start {
			break
		}
		implement := element.Value.(func(Arg) error)
		err = implement(arg)
		if err == nil {
			this.last = element
			return nil
		}
		element = element.Next()
	}
	return err
}

func (this *ServiceResult[Result]) Register(implement func() (Result, error)) func() {
	if this.implements == nil {
		this.implements = list.New()
	}
	element := this.implements.PushBack(implement)
	return func() {
		this.implements.Remove(element)
	}
}

func (this *ServiceResult[Result]) Call() (Result, error) {
	var result Result
	if this.implements == nil {
		return result, NotImplement
	}
	start := this.last
	if start == nil || start.Prev() == nil {
		start = this.implements.Front()
	}
	err := NotImplement
	element := start
	for {
		if element == nil {
			element = this.implements.Front()
		}
		if element == start {
			break
		}
		implement := element.Value.(func() (Result, error))
		result, err = implement()
		if err == nil {
			this.last = element
			return result, nil
		}
		element = element.Next()
	}
	return result, err
}

func (this *ServiceArgResult[Arg, Result]) Register(implement func(Arg) (Result, error)) func() {
	if this.implements == nil {
		this.implements = list.New()
	}
	element := this.implements.PushBack(implement)
	return func() {
		this.implements.Remove(element)
	}
}

func (this *ServiceArgResult[Arg, Result]) Call(arg Arg) (Result, error) {
	var result Result
	if this.implements == nil {
		return result, NotImplement
	}
	start := this.last
	if start == nil || start.Prev() == nil {
		start = this.implements.Front()
	}
	err := NotImplement
	element := start
	for {
		if element == nil {
			element = this.implements.Front()
		}
		if element == start {
			break
		}
		implement := element.Value.(func(Arg) (Result, error))
		result, err = implement(arg)
		if err == nil {
			this.last = element
			return result, nil
		}
		element = element.Next()
	}
	return result, err
}

var NotImplement = errors.New("not implement")
