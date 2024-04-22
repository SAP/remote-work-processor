package processors

import (
	"context"
	"fmt"
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
	ctx := executors.NewExecutorContext(p.req.GetInput(), p.req.Store)

	if !p.isEnabled() {
		log.Println("Unable to process remote task. Remote Worker is disabled...")
		return &pb.ClientMessage{
			Body: buildResult(ctx, p.req, executors.NewExecutorResult(
				executors.Status(pb.TaskExecutionResponseMessage_TASK_STATE_FAILED_NON_CHARGEABLE),
				executors.Error(fmt.Errorf("unable to process remote task. Remote Worker is disabled")),
			)),
		}, nil
	}

	log.Println("Processing Task...")
	executor := factory.CreateExecutor(p.req.GetType())

	res := executor.Execute(ctx)
	return &pb.ClientMessage{
		Body: buildResult(ctx, p.req, res),
	}, nil
}

func buildResult(ctx executors.Context, req *pb.TaskExecutionRequestMessage, res *executors.ExecutorResult) *pb.ClientMessage_TaskExecutionResponse {
	return &pb.ClientMessage_TaskExecutionResponse{
		TaskExecutionResponse: &pb.TaskExecutionResponseMessage{
			ExecutionId:      req.GetExecutionId(),
			ExecutionVersion: req.GetExecutionVersion(),
			State:            res.Status,
			Output:           res.Output,
			Store:            ctx.GetStore(),
			Error: &wrapperspb.StringValue{
				Value: res.Error,
			},
			Type: req.Type,
		},
	}
}
