package processors

import (
	"context"
	"log"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/executors"
	"github.com/SAP/remote-work-processor/internal/executors/factory"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type RemoteTaskProcessor struct {
	req       *pb.TaskExecutionRequestMessage
	isEnabled func() bool
}

func NewRemoteTaskProcessor(req *pb.ServerMessage_TaskExecutionRequest, isEnabled func() bool) RemoteTaskProcessor {
	return RemoteTaskProcessor{
		req:       req.TaskExecutionRequest,
		isEnabled: isEnabled,
	}
}

func (p RemoteTaskProcessor) Process(_ context.Context) (*pb.ClientMessage, error) {
	if !p.isEnabled() {
		log.Println("Unable to process remote task. Remote Worker is disabled...")
		return nil, nil
	}

	log.Println("Processing Task...")
	executor, err := factory.CreateExecutor(p.req.GetType())
	if err != nil {
		log.Println(err)
		// Do not fail and recreate gRPC connection on unsupported task type
		return nil, nil
	}

	ctx := executors.NewExecutorContext(p.req.GetInput(), p.req.Store)

	res := executor.Execute(ctx)
	return &pb.ClientMessage{
		Body: buildResult(p.req, res),
	}, nil
}

func buildResult(req *pb.TaskExecutionRequestMessage, res *executors.ExecutorResult) *pb.ClientMessage_TaskExecutionResponse {
	return &pb.ClientMessage_TaskExecutionResponse{
		TaskExecutionResponse: &pb.TaskExecutionResponseMessage{
			ExecutionId:      req.GetExecutionId(),
			ExecutionVersion: req.GetExecutionVersion(),
			State:            res.Status,
			Output:           res.Output,
			Store:            res.Store,
			Error: &wrapperspb.StringValue{
				Value: res.Error,
			},
			Type: req.Type,
		},
	}
}
