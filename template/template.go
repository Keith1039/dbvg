// Package template contains all the templating logic
//
// the validate package contains all functions that relate to creating or validating templates
package template

import (
	"fmt"
	"github.com/Keith1039/dbvg/strategy"
	"github.com/Keith1039/dbvg/utils"
)

type Template interface {
	TemplateFrom(string) error // parses a template file and validates it's contents, it then fills the struct
}

// map the key to a value
func makeTypeMap(data map[string]any) map[string]string {
	m := make(map[string]string)
	for key, val := range data {
		m[key] = fmt.Sprintf("%T", val)
	}
	return m
}

// check if the type map matches the given schema
func checkAgainstSchema(typeMap map[string]string, schema map[string]string) error {
	if len(typeMap) != len(schema) {
		return SchemaError{expectedSchema: schema, actualSchema: typeMap}
	} else {
		for key, val := range typeMap {
			// check the schema values but make an exception for any
			if val != schema[key] && schema[key] != "any" {
				return SchemaError{expectedSchema: schema, actualSchema: typeMap}
			}
		}
	}
	return nil
}

func normalizeKeys(columnInfo map[string]any) {
	var normalizedKey string
	keysToDelete := make(map[string]bool)

	for key, val := range columnInfo {
		normalizedKey = utils.TrimAndLowerString(key)
		if normalizedKey != key {
			columnInfo[normalizedKey] = val
			keysToDelete[key] = true
		}
	}

	for key := range keysToDelete {
		delete(columnInfo, key)
	}
}

func normalizeType(columnInfo map[string]any) {
	columnInfo["type"] = utils.TrimAndUpperString(columnInfo["type"].(string))
}
func wrapError(tableName string, columnName string, err error) error {
	return fmt.Errorf("for column '%s' in table '%s': [%w]", columnName, tableName, err)
}

// checks if the expected type string matches the received type
func checkExpectedType(expectedType string, receivedType string) error {
	// behavior should change with config, if lax, this should give a warning and coerce the received type to expected and log this transformation
	// if strict, return a genuine error
	if expectedType != receivedType {
		return strategy.UnexpectedTypeError{ExpectedType: expectedType, ActualType: receivedType}
	}
	return nil
}
