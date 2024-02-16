package executors

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/SAP/remote-work-processor/internal/cache"
)

type Context interface {
	GetString(k string) string
	GetRequiredString(k string) (string, error)
	GetNumber(k string) (uint64, error)
	GetRequiredNumber(k string) (uint64, error)
	GetMap(k string) (map[string]string, error)
	GetList(k string) ([]string, error)
	GetBoolean(k string) (bool, error)
	GetStore() cache.MapCache[string, string]
}

type ExecutorContext struct {
	input map[string]string
	store cache.MapCache[string, string]
}

var (
	bools = map[string]bool{
		"true":  true,
		"false": false,
	}
)

func NewExecutorContext(input map[string]string, store map[string]string) ExecutorContext {
	return ExecutorContext{
		input: input,
		store: cache.NewInMemoryCache[string, string]().FromMap(store),
	}
}

func (e *ExecutorContext) GetString(k string) string {
	return e.input[k]
}

func (e *ExecutorContext) GetRequiredString(k string) (string, error) {
	if v, ok := e.input[k]; ok {
		return v, nil
	}

	return "", NewRequiredKeyValidationError(k)
}

func (e *ExecutorContext) GetNumber(k string) (uint64, error) {
	s, ok := e.input[k]
	if !ok {
		return 0, nil
	}

	n, err := strconv.ParseUint(s, 10, 64)
	if errors.Is(&strconv.NumError{}, err) {
		return 0, err
	}

	return n, nil
}

func (e *ExecutorContext) GetRequiredNumber(k string) (uint64, error) {
	if _, ok := e.input[k]; ok {
		return e.GetNumber(k)
	}

	return 0, NewRequiredKeyValidationError(k)
}

func (e *ExecutorContext) GetMap(k string) (map[string]string, error) {
	m := make(map[string]string)
	s, ok := e.input[k]
	if !ok {
		return m, nil
	}

	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil, err
	}

	return m, nil
}

func (e *ExecutorContext) GetList(k string) ([]string, error) {
	l := make([]string, 0)
	s, ok := e.input[k]
	if !ok {
		return l, nil
	}

	if err := json.Unmarshal([]byte(s), &l); err != nil {
		return nil, err
	}

	return l, nil
}

func (e *ExecutorContext) GetBoolean(k string) (bool, error) {
	s, ok := e.input[k]
	if !ok {
		return false, nil
	}

	b, ok := bools[s]
	if !ok {
		return false, NewNonRetryableError("Input value %q for key %q is not a valid boolean", s, k)
	}

	return b, nil
}

func (e *ExecutorContext) GetStore() cache.MapCache[string, string] {
	return e.store
}
