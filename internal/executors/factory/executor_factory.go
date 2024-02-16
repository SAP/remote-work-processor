package factory

import (
	"fmt"
	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/executors"
	"github.com/SAP/remote-work-processor/internal/executors/http"
	"github.com/SAP/remote-work-processor/internal/executors/void"
)

func CreateExecutor(t pb.TaskType) (executors.Executor, error) {
	switch t {
	case pb.TaskType_TASK_TYPE_VOID:
		return void.VoidExecutor{}, nil
	case pb.TaskType_TASK_TYPE_HTTP:
		return http.NewDefaultHttpRequestExecutor(), nil
	default:
		return nil, fmt.Errorf("cannot create executor of type %q", t)
	}
}
