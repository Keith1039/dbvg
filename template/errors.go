package template

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SchemaError is an error given when the given schema doesn't match the expected schema
type SchemaError struct {
	expectedSchema map[string]string
	actualSchema   map[string]string
}

func (err SchemaError) Error() string {
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

// UndefinedDefaultError is an error given when a supported type does not have a default code associated with it
type UndefinedDefaultError struct {
	columnType string
}

func (err UndefinedDefaultError) Error() string {
	return fmt.Sprintf("undefined default for column type: '%s'", err.columnType)
}

type MissingRequiredTableError struct {
	tableName string
	jsonKeys  []string
}

func (err MissingRequiredTableError) Error() string {
	return fmt.Sprintf("missing required table '%s' in template keys [%s]", err.tableName, strings.Join(err.jsonKeys, ", "))
}
