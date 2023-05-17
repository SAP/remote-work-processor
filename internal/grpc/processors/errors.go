package processors

import "fmt"

type ProcessorError struct {
	message string
}

func NewProcessorError(message string) *ProcessorError {
	return &ProcessorError{
		message: message,
	}
}

func (e ProcessorError) Error() string {
	if e.message == "" {
		return fmt.Sprint(e.message)
	}

	return fmt.Sprintf("An error occurred while processing operation")
}
