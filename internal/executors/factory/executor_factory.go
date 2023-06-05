package factory

import (
	"log"
	"sync"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
	"github.com/SAP/remote-work-processor/internal/executors"
)

var (
	Executor_Factory ExecutorFactory = newExecutorFactory()
)

type ExecutorFactory struct {
	generators map[pb.TaskType]ExecutorGenerator
	sync.RWMutex
}

func newExecutorFactory() ExecutorFactory {
	return ExecutorFactory{
		generators: map[pb.TaskType]ExecutorGenerator{
			pb.TaskType_TASK_TYPE_VOID:                   voidExecutorGenerator(),
			pb.TaskType_TASK_TYPE_HTTP:                   httpRequestExecutorGenerator(),
			pb.TaskType_TASK_TYPE_KUBERNETES_API_REQUEST: kubernetesApiRequestExecutorGenerator(),
		},
	}
}

func (f *ExecutorFactory) Submit(t pb.TaskType, g ExecutorGenerator) *ExecutorFactory {
	f.Lock()
	defer f.Unlock()

	if _, e := f.generators[t]; e {
		log.Fatalf("Executor of type '%s' has already been submitted in the factory", t)
	}

	f.generators[t] = g
	return f
}

func (f *ExecutorFactory) GetExecutor(t pb.TaskType) (executors.Executor, error) {
	f.RLock()
	g, ok := f.generators[t]
	f.RUnlock()

	if !ok {
		return nil, executors.NewExecutorCreationError(t)
	}

	e, err := g()
	if err != nil {
		log.Fatalf("Generator failed while trying to create an executor of type '%s'", t)
	}

	return e, nil
}
