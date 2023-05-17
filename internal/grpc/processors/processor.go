package processors

import (
	pb "github.com/SAP/remote-work-processor/build/proto/generated"
)

type Processor interface {
	Process() <-chan *ProcessorResult
}
type ProcessorResult struct {
	Result *pb.ClientMessage
	Err    error
	Done   chan struct{}
}

type processorResultOption func(*ProcessorResult)

func NewProcessorResult(opts ...processorResultOption) *ProcessorResult {
	pr := &ProcessorResult{}

	for _, opt := range opts {
		opt(pr)
	}

	return pr
}

func Result(r *pb.ClientMessage) processorResultOption {
	return func(pr *ProcessorResult) {
		pr.Result = r
	}
}

func Error(err error) processorResultOption {
	return func(pr *ProcessorResult) {
		pr.Err = err
	}
}

func OnChannel(done chan struct{}) processorResultOption {
	return func(pr *ProcessorResult) {
		pr.Done = done
	}
}
