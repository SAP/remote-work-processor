package executors

type ExecutorType uint

const (
	ExecutorType_UNKNOWN ExecutorType = iota
	ExecutorType_VOID
	ExecutorType_HTTP
	ExecutorType_SCRIPT
)

var (
	executorTypeNames = [...]string{"VOID", "HTTP"}
)

func (t ExecutorType) String() string {
	return executorTypeNames[t]
}

func (e ExecutorType) Ordinal() uint {
	return uint(e)
}
