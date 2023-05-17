package processors

import (
	"encoding/json"
	"log"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/executors"
	"github.com/SAP/remote-work-processor/internal/executors/factory"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type RemoteTaskProcessor struct {
	req *pb.ServerMessage_TaskExecutionRequest
}

func NewRemoteTaskProcessor(req *pb.ServerMessage_TaskExecutionRequest) RemoteTaskProcessor {
	return RemoteTaskProcessor{
		req: req,
	}
}

func (p RemoteTaskProcessor) Process() <-chan *ProcessorResult {
	c := make(chan *ProcessorResult)
	go func() {
		log.Println("Processing remote task...")
		executor, err := factory.Executor_Factory.GetExecutor(p.req.TaskExecutionRequest.GetType())
		if err != nil {
			c <- NewProcessorResult(Error(err))
		}

		ctx := executors.NewExecutorContext(p.req.TaskExecutionRequest.GetInput(), p.req.TaskExecutionRequest.Store)

		res := executor.Execute(ctx)
		c <- NewProcessorResult(Result(&pb.ClientMessage{
			Body: buildResult(ctx, p.req, res),
		}))
	}()

	return c
}

func buildResult(ctx executors.ExecutorContext, req *pb.ServerMessage_TaskExecutionRequest, res *executors.ExecutorResult) *pb.ClientMessage_TaskExecutionResponse {
	return &pb.ClientMessage_TaskExecutionResponse{
		TaskExecutionResponse: &pb.TaskExecutionResponseMessage{
			ExecutionId:      req.TaskExecutionRequest.GetExecutionId(),
			ExecutionVersion: req.TaskExecutionRequest.GetExecutionVersion(),
			State:            res.Status,
			Output:           toStringValues(res.Output),
			Store:            ctx.GetStore().ToMap(),
			Error: &wrapperspb.StringValue{
				Value: res.Error,
			},
			Type: req.TaskExecutionRequest.Type,
		},
	}
}

func toStringValues(m map[string]interface{}) map[string]string {
	out := make(map[string]string)
	for k, v := range m {
		if str, ok := v.(string); ok {
			out[k] = str
			continue
		}

		b, err := json.Marshal(v)
		if err != nil {
			log.Fatalf("Failed to serialize value %s: %v", v, err)
		}
		out[k] = string(b)
	}

	return out
}
