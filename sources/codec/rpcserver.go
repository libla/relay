package codec

type ResponseContext[TRequest, TResponse any] interface {
	Context
	Next(TRequest, TResponse) error
}

type ResponseHandle[TInput, TOutput any] interface {
	Handle(TInput, func(TOutput, error))
}

type Response[TInput, TOutput, TRequest, TResponse any] interface {
	Call(ResponseContext[TRequest, TResponse], TRequest, ResponseHandle[TInput, TOutput]) error
}

type ResponseFunc[TInput, TOutput any] func(TInput, func(TOutput, error))

func (this ResponseFunc[TInput, TOutput]) Handle(input TInput, output func(TOutput, error)) {
	this(input, output)
}

func LinkResponse[TInput, TOutput, TRequest, TResponse, TBetweenIn, TBetweenOut any](response Response[TBetweenIn, TBetweenOut, TRequest, TResponse], input Decoder[TBetweenIn, TInput], output Encoder[TBetweenOut, TOutput]) Response[TInput, TOutput, TRequest, TResponse] {
	return &linkResponse[TInput, TOutput, TRequest, TResponse, TBetweenIn, TBetweenOut]{decoder: input, encoder: output, response: response}
}

type linkResponse[TInput, TOutput, TRequest, TResponse, TBetweenIn, TBetweenOut any] struct {
	decoder  Decoder[TBetweenIn, TInput]
	encoder  Encoder[TBetweenOut, TOutput]
	response Response[TBetweenIn, TBetweenOut, TRequest, TResponse]
}

func (this *linkResponse[TInput, TOutput, TRequest, TResponse, TBetweenIn, TBetweenOut]) Call(context ResponseContext[TRequest, TResponse], input TRequest, handle ResponseHandle[TInput, TOutput]) error {
	link := linkResponseHandle[TInput, TOutput, TRequest, TResponse, TBetweenIn, TBetweenOut]{
		decoder: this.decoder,
		in: linkResponseInContext[TInput, TOutput, TRequest, TResponse, TBetweenOut]{
			ResponseContext: context, output: this.encoder, handle: handle,
			out: linkResponseOutContext[TRequest, TResponse, TBetweenOut]{ResponseContext: context},
		},
	}
	err := this.response.Call(context, input, &link)
	if err != nil {
		return err
	}
	return link.err
}

type linkResponseHandle[TInput, TOutput, TRequest, TResponse, TBetweenIn, TBetweenOut any] struct {
	decoder Decoder[TBetweenIn, TInput]
	in      linkResponseInContext[TInput, TOutput, TRequest, TResponse, TBetweenOut]
	err     error
}

func (this *linkResponseHandle[TInput, TOutput, TRequest, TResponse, TBetweenIn, TBetweenOut]) Handle(input TBetweenIn, output func(TBetweenOut, error)) {
	this.in.out.result = output
	err := this.decoder.Decode(&this.in, input)
	if err != nil {
		this.err = err
	}
}

type linkResponseInContext[TInput, TOutput, TRequest, TResponse, TBetween any] struct {
	ResponseContext[TRequest, TResponse]
	output Encoder[TBetween, TOutput]
	handle ResponseHandle[TInput, TOutput]
	out    linkResponseOutContext[TRequest, TResponse, TBetween]
}

func (this *linkResponseInContext[TInput, TOutput, TRequest, TResponse, TBetween]) Next(input TInput) error {
	this.handle.Handle(input, func(output TOutput, err error) {
		var result TBetween
		if err != nil {
			this.out.result(result, err)
		} else {
			err = this.output.Encode(&this.out, output)
			if err != nil {
				this.out.result(result, err)
			}
		}
	})
	return nil
}

type linkResponseOutContext[TRequest, TResponse, TResult any] struct {
	ResponseContext[TRequest, TResponse]
	result func(TResult, error)
}

func (this *linkResponseOutContext[TRequest, TResponse, TResult]) Next(value TResult) error {
	this.result(value, nil)
	return nil
}
