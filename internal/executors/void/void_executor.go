package void

import (
	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/executors"
)

const (
	MESSAGE_KEY = "message"
)

type VoidExecutor struct{}

func (VoidExecutor) Execute(ctx executors.ExecutorContext) *executors.ExecutorResult {
	msg := ctx.GetString(MESSAGE_KEY)
	return executors.NewExecutorResult(
		executors.Output(buildOutput(msg)),
		executors.Status(pb.TaskExecutionResponseMessage_TASK_STATE_COMPLETED),
	)
}

func buildOutput(msg string) map[string]any {
	return map[string]any{
		MESSAGE_KEY: msg,
	}
}
