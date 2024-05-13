package factory

import (
	"fmt"
	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/executors"
	"github.com/SAP/remote-work-processor/internal/executors/http"
	"github.com/SAP/remote-work-processor/internal/executors/void"
)

type errorExecutor func() error

func CreateExecutor(t pb.TaskType) executors.Executor {
	switch t {
	case pb.TaskType_TASK_TYPE_VOID:
		return void.VoidExecutor{}
	case pb.TaskType_TASK_TYPE_HTTP:
		return http.NewDefaultHttpRequestExecutor()
	default:
		return errorExecutor(func() error {
			return fmt.Errorf("cannot create executor of type %q: unsupported task type", t)
		})
	}
}

func (exec errorExecutor) Execute(_ executors.Context) *executors.ExecutorResult {
	return executors.NewExecutorResult(
		executors.Status(pb.TaskExecutionResponseMessage_TASK_STATE_FAILED_NON_CHARGEABLE),
		executors.Error(exec()),
	)
}
