package factory

import (
	"github.com/SAP/remote-work-processor/internal/executors"
	"github.com/SAP/remote-work-processor/internal/executors/http"
	"github.com/SAP/remote-work-processor/internal/executors/kubernetes"
	"github.com/SAP/remote-work-processor/internal/executors/void"
)

type ExecutorGenerator func() (executors.Executor, error)

func voidExecutorGenerator() ExecutorGenerator {
	return func() (executors.Executor, error) {
		return &void.VoidExecutor{}, nil
	}
}

func httpRequestExecutorGenerator() ExecutorGenerator {
	return func() (executors.Executor, error) {
		return &http.HttpRequestExecutor{}, nil
	}
}

func kubernetesApiRequestExecutorGenerator() ExecutorGenerator {
	return func() (executors.Executor, error) {
		return &kubernetes.KubernetesApiRequestExecutor{}, nil
	}
}
