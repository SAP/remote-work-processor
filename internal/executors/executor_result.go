package executors

import pb "github.com/SAP/remote-work-processor/build/proto/generated"

type ExecutorResult struct {
	Output map[string]string
	Status pb.TaskExecutionResponseMessage_TaskState
	Error  string
}

type ExecutorResultOption func(*ExecutorResult)

func NewExecutorResult(opts ...ExecutorResultOption) *ExecutorResult {
	r := &ExecutorResult{
		Output: make(map[string]string),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func Output(m map[string]string) ExecutorResultOption {
	return func(er *ExecutorResult) {
		for key, val := range m {
			er.Output[key] = val
		}
	}
}

func Status(s pb.TaskExecutionResponseMessage_TaskState) ExecutorResultOption {
	return func(er *ExecutorResult) {
		er.Status = s
	}
}

func Error(err error) ExecutorResultOption {
	return func(er *ExecutorResult) {
		if err != nil {
			er.Error = err.Error()
		}
	}
}

func ErrorString(err string) ExecutorResultOption {
	return func(er *ExecutorResult) {
		er.Error = err
	}
}
