package void

import (
	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/executors"
	"log"
)

const (
	MESSAGE_KEY = "message"
)

type VoidExecutor struct{}

func (VoidExecutor) Execute(ctx executors.ExecutorContext) *executors.ExecutorResult {
	log.Println("Executing Void command...")
	msg := ctx.GetString(MESSAGE_KEY)
	return executors.NewExecutorResult(
		executors.Output(buildOutput(msg)),
		executors.Status(pb.TaskExecutionResponseMessage_TASK_STATE_COMPLETED),
	)
}

func buildOutput(msg string) map[string]string {
	return map[string]string{
		MESSAGE_KEY: msg,
	}
}
