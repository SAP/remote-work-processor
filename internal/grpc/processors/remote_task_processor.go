package processors

import (
	"context"
	"encoding/json"
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

	log.Println("Processing remote task...")
	executor, err := factory.CreateExecutor(p.req.GetType())
	if err != nil {
		return nil, err
	}

	ctx := executors.NewExecutorContext(p.req.GetInput(), p.req.Store)

	res := executor.Execute(ctx)
	return &pb.ClientMessage{
		Body: buildResult(ctx, p.req, res),
	}, nil
}

func buildResult(ctx executors.ExecutorContext, req *pb.TaskExecutionRequestMessage,
	res *executors.ExecutorResult) *pb.ClientMessage_TaskExecutionResponse {
	return &pb.ClientMessage_TaskExecutionResponse{
		TaskExecutionResponse: &pb.TaskExecutionResponseMessage{
			ExecutionId:      req.GetExecutionId(),
			ExecutionVersion: req.GetExecutionVersion(),
			State:            res.Status,
			Output:           toStringValues(res.Output),
			Store:            ctx.GetStore(), // FIXME: this returns the store from the request, not the processed one
			Error: &wrapperspb.StringValue{
				Value: res.Error,
			},
			Type: req.Type,
		},
	}
}

func toStringValues(m map[string]interface{}) map[string]string {
	result := make(map[string]string, len(m))
	for key, value := range m {
		if str, ok := value.(string); ok {
			result[key] = str
			continue
		}

		serialized, err := json.Marshal(value)
		if err != nil {
			log.Printf("Failed to serialize value %q: %v", value, err)
			serialized = []byte("<nil>")
		}
		result[key] = string(serialized)
	}
	return result
}
