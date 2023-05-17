package executors

import (
	"fmt"
	"log"

	pb "github.com/SAP/remote-work-processor/build/proto/generated"
)

type RequiredKeyValidationError struct {
	key string
}

func NewRequiredKeyValidationError(key string) *RequiredKeyValidationError {
	if len(key) == 0 {
		log.Fatal("Key cannot be blank")
	}

	return &RequiredKeyValidationError{
		key: key,
	}
}

func (err *RequiredKeyValidationError) Error() string {
	return fmt.Sprintf("Key '%s' is required but it had not been provided", err.key)
}

type ExecutorCreationError struct {
	t pb.TaskType
}

func NewExecutorCreationError(t pb.TaskType) *ExecutorCreationError {
	return &ExecutorCreationError{
		t: t,
	}
}

func (err *ExecutorCreationError) Error() string {
	return fmt.Sprintf("Cannot create executor of type '%s'", err.t)
}

type NonRetryableError struct {
	msg   string
	cause error
}

func NewNonRetryableError(msg string) *NonRetryableError {
	return &NonRetryableError{
		msg: msg,
	}
}

func (err *NonRetryableError) WithCause(e error) *NonRetryableError {
	err.cause = e
	return err
}

func (err *NonRetryableError) Error() string {
	return err.msg
}

func (err *NonRetryableError) Unwrap() error {
	return err.cause
}

type RetryableError struct {
	msg string
	err error
}

func NewRetryableError(msg string) *RetryableError {
	return &RetryableError{
		msg: msg,
	}
}

func (err *RetryableError) WithCause(e error) *RetryableError {
	err.err = e
	return err
}

func (err *RetryableError) Error() string {
	return err.msg
}

func (err *RetryableError) Unwrap() error {
	return err.err
}

type InvalidHttpMethodError struct {
	m string
}

func NewInvalidHttpMethodError(m string) *InvalidHttpMethodError {
	return &InvalidHttpMethodError{
		m: m,
	}
}

func (err *InvalidHttpMethodError) Error() string {
	return fmt.Sprintf("'%s' is not a valid HTTP method.", err.m)
}
