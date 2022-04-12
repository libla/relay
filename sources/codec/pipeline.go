package codec

type PipelineContext[TValue any] interface {
	Context
	Next(TValue) error
}

type Encoder[TInput, TOutput any] interface {
	Encode(PipelineContext[TInput], TOutput) error
}

type Decoder[TInput, TOutput any] interface {
	Decode(PipelineContext[TOutput], TInput) error
}

type Pipeline[TInput, TOutput any] interface {
	Encoder[TInput, TOutput]
	Decoder[TInput, TOutput]
}

type PassPipeline[T any] struct{}

func (this PassPipeline[T]) Encode(context PipelineContext[T], output T) error {
	return context.Next(output)
}

func (this PassPipeline[T]) Decode(context PipelineContext[T], input T) error {
	return context.Next(input)
}

func LinkPipeline[TInput, TOutput, TBetween any](input Pipeline[TInput, TBetween], output Pipeline[TBetween, TOutput]) Pipeline[TInput, TOutput] {
	return &linkPipeline[TInput, TOutput, TBetween]{input: input, output: output}
}

type linkPipeline[TInput, TOutput, TBetween any] struct {
	input  Pipeline[TInput, TBetween]
	output Pipeline[TBetween, TOutput]
}

func (this *linkPipeline[TInput, TOutput, TBetween]) Encode(context PipelineContext[TInput], output TOutput) error {
	return this.output.Encode(&linkEncodeContext[TInput, TBetween]{context, this.input}, output)
}

func (this *linkPipeline[TInput, TOutput, TBetween]) Decode(context PipelineContext[TOutput], input TInput) error {
	return this.input.Decode(&linkDecodeContext[TOutput, TBetween]{context, this.output}, input)
}

type linkEncodeContext[TInput, TOutput any] struct {
	PipelineContext[TInput]
	next Pipeline[TInput, TOutput]
}

type linkDecodeContext[TInput, TOutput any] struct {
	PipelineContext[TInput]
	next Pipeline[TOutput, TInput]
}

func (this *linkEncodeContext[TInput, TOutput]) Next(value TOutput) error {
	return this.next.Encode(this.PipelineContext, value)
}

func (this *linkDecodeContext[TInput, TOutput]) Next(value TOutput) error {
	return this.next.Decode(this.PipelineContext, value)
}
