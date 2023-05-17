package executors

type ExecutorStatus uint

const (
	ExecutorStatus_UNKNOWN ExecutorStatus = iota
	ExecutorStatus_COMPLETED
	ExecutorStatus_FAILED_RETRYABLE
	ExecutorStatus_FAILED_NON_RETRYABLE
)

var (
	executorStatusNames = [...]string{"COMPLETED", "FAILED_RETRYABLE", "FAILED_NON_RETRYABLE"}
)

func (es ExecutorStatus) String() string {
	return executorStatusNames[es]
}

func (es ExecutorStatus) Ordinal() uint {
	return uint(es)
}
