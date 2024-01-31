package executors

import "fmt"

type RequiredKeyValidationError string

func NewRequiredKeyValidationError(key string) error {
	return RequiredKeyValidationError(key)
}

func (err RequiredKeyValidationError) Error() string {
	return fmt.Sprintf("key %q is required but not provided", string(err))
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
