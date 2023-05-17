package executors

type Executor interface {
	Execute(ctx ExecutorContext) *ExecutorResult
}
