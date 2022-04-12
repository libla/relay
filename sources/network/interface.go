package network

type Session[TValue any] interface {
	Start() error
	Close() error
	Connected() bool
	Started() bool
	Send(TValue) error
}

type SessionHandle[TInput, TOutput any] interface {
	OnClose(Session[TOutput]) error
	OnMessage(Session[TOutput], TInput) error
	OnError(Session[TOutput], error)
}

type ListenerHandle[TInput, TOutput any] interface {
	SessionHandle[TInput, TOutput]
	OnAccept(Session[TOutput]) error
}

func NewSessionHandle[TInput, TOutput any](message func(Session[TOutput], TInput) error, close func(Session[TOutput]) error, error func(Session[TOutput], error)) SessionHandle[TInput, TOutput] {
	return &sessionHandle[TInput, TOutput]{message, close, error}
}

func NewListenerHandle[TInput, TOutput any](message func(Session[TOutput], TInput) error, accept func(Session[TOutput]) error, close func(Session[TOutput]) error, error func(Session[TOutput], error)) ListenerHandle[TInput, TOutput] {
	return &listenerHandle[TInput, TOutput]{sessionHandle[TInput, TOutput]{message, close, error}, accept}
}

type sessionHandle[TInput, TOutput any] struct {
	message func(Session[TOutput], TInput) error
	close   func(Session[TOutput]) error
	error   func(Session[TOutput], error)
}

type listenerHandle[TInput, TOutput any] struct {
	sessionHandle[TInput, TOutput]
	accept func(Session[TOutput]) error
}

func (this *sessionHandle[TInput, TOutput]) OnClose(session Session[TOutput]) error {
	if this.close != nil {
		return this.close(session)
	}
	return nil
}

func (this *sessionHandle[TInput, TOutput]) OnMessage(session Session[TOutput], input TInput) error {
	if this.message != nil {
		return this.message(session, input)
	}
	return nil
}

func (this *sessionHandle[TInput, TOutput]) OnError(session Session[TOutput], err error) {
	if this.error != nil {
		this.error(session, err)
	} else {
		panic(err)
	}
}

func (this *listenerHandle[TInput, TOutput]) OnAccept(session Session[TOutput]) error {
	if this.accept != nil {
		return this.accept(session)
	}
	session.Start()
	return nil
}
