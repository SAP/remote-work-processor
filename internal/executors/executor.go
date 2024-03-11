package executors

type Executor interface {
	Execute(Context) *ExecutorResult
}
