package json

import (
	"encoding/json"
)

func ToJson[T any](value T) (string, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func FromJson[S ~string, R any](serializable S, destination *R) error {
	if err := json.Unmarshal([]byte(serializable), destination); err != nil {
		return err
	}

	return nil
}
