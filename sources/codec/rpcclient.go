package codec

type RequestContext[TRequest, TResponse any] interface {
	Context
	Next(TRequest) (TResponse, error)
}

type Request[TInput, TOutput, TRequest, TResponse any] interface {
	Call(context RequestContext[TRequest, TResponse], request TInput) (TOutput, error)
}

func LinkRequest[TInput, TOutput, TRequest, TResponse, TBetweenIn, TBetweenOut any](request Request[TBetweenIn, TBetweenOut, TRequest, TResponse], input Encoder[TBetweenIn, TInput], output Decoder[TBetweenOut, TOutput]) Request[TInput, TOutput, TRequest, TResponse] {
	return &linkRequest[TInput, TOutput, TRequest, TResponse, TBetweenIn, TBetweenOut]{encoder: input, decoder: output, request: request}
}

type linkRequest[TInput, TOutput, TRequest, TResponse, TBetweenIn, TBetweenOut any] struct {
	encoder Encoder[TBetweenIn, TInput]
	decoder Decoder[TBetweenOut, TOutput]
	request Request[TBetweenIn, TBetweenOut, TRequest, TResponse]
}

func (this *linkRequest[TInput, TOutput, TRequest, TResponse, TBetweenIn, TBetweenOut]) Call(context RequestContext[TRequest, TResponse], request TInput) (TOutput, error) {
	in := linkRequestInContext[TOutput, TRequest, TResponse, TBetweenIn, TBetweenOut]{
		RequestContext: context, decoder: this.decoder, request: this.request, out: linkRequestOutContext[TRequest, TResponse, TOutput]{RequestContext: context},
	}
	err := this.encoder.Encode(&in, request)
	return in.out.result, err
}

type linkRequestInContext[TOutput, TRequest, TResponse, TBetweenIn, TBetweenOut any] struct {
	RequestContext[TRequest, TResponse]
	decoder Decoder[TBetweenOut, TOutput]
	request Request[TBetweenIn, TBetweenOut, TRequest, TResponse]
	out     linkRequestOutContext[TRequest, TResponse, TOutput]
}

func (this *linkRequestInContext[TOutput, Request, TResponse, TBetweenIn, TBetweenOut]) Next(value TBetweenIn) error {
	result, err := this.request.Call(this.RequestContext, value)
	if err != nil {
		return err
	}
	return this.decoder.Decode(&this.out, result)
}

type linkRequestOutContext[TRequest, TResponse, TResult any] struct {
	RequestContext[TRequest, TResponse]
	result TResult
}

func (this *linkRequestOutContext[TRequest, TResponse, TResult]) Next(value TResult) error {
	this.result = value
	return nil
}
