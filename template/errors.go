package template

import (
	"encoding/json"
	"fmt"
)

// schemaError is an error given when the given schema doesn't match the expected schema
type schemaError struct {
	expectedSchema map[string]string
	actualSchema   map[string]string
}

func (err schemaError) Error() string {
	data, _ := json.MarshalIndent(err.expectedSchema, "", " ") // assuming schema data isn't nil
	data2, _ := json.MarshalIndent(err.actualSchema, "", " ")
	return fmt.Sprintf("expected schema:\n\t%s\n, received:\n\t%s", string(data), string(data2))
}

// PreprocessError is an error given when a value is of type []any but couldn't be processed into []int, []float64 or []string
type PreprocessError struct {
	val any
}

func (err PreprocessError) Error() string {
	return fmt.Sprintf("could not preprocess value '%v' into []int, []float64 or []string", err.val)
}
