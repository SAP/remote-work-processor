package executors

import pb "github.com/SAP/remote-work-processor/build/proto/generated"

type ExecutorResult struct {
	Output map[string]any
	Status pb.TaskExecutionResponseMessage_TaskState
	Error  string
}

type executorResultOption func(*ExecutorResult)

func NewExecutorResult(opts ...executorResultOption) *ExecutorResult {
	r := &ExecutorResult{}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func Output(m map[string]any) executorResultOption {
	return func(er *ExecutorResult) {
		er.Output = m
	}
}

func Status(s pb.TaskExecutionResponseMessage_TaskState) executorResultOption {
	return func(er *ExecutorResult) {
		er.Status = s
	}
}

func Error(err error) executorResultOption {
	return func(er *ExecutorResult) {
		if err == nil {
			return
		}

		er.Error = err.Error()
	}
}

func ErrorString(err string) executorResultOption {
	return func(er *ExecutorResult) {
		er.Error = err
	}
}
