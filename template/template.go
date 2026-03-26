// Package template contains all the templating logic
//
// the validate package contains all functions that relate to creating or validating templates
package template

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Keith1039/dbvg/strategy"
	"github.com/Keith1039/dbvg/utils"
	"os"
)

// NOTE: This section of code will eventually be moved to parsers or something that parsers use
// NULL isn't included since it's by default always supported
var overrideCodes = map[string]map[string]bool{
	"DATE": {"NOW": true},
	"UUID": {"UUID": true},
	"BOOL": {"RANDOM": true},
	"VARCHAR": {
		"EMAIL":     true,
		"FIRSTNAME": true,
		"LASTNAME":  true,
		"FULLNAME":  true,
		"PHONE":     true,
		"COUNTRY":   true,
		"ADDRESS":   true,
		"ZIPCODE":   true,
		"CITY":      true,
	},
}

// codes that don't need a value but can still be given one
var optionalCodes = map[string]map[string]bool{
	"INT": {"SEQ": true},
}

// codes that use a value (if none is given default is configured in parser
var requiredCodeMap = map[string]map[string]bool{
	"BOOL": {
		"STATIC": true,
	},
	"DATE": {
		"RANDOM": true,
		"STATIC": true,
	},
	"INT": {
		"RANDOM": true,
		"STATIC": true,
	},
	"FLOAT": {
		"RANDOM": true,
		"STATIC": true,
	},
	"VARCHAR": {
		"STATIC": true,
		"REGEX":  true,
	},
}

// section end

type Template interface {
	TemplateFrom(string) error // parses a template file and validates it's contents, it then fills the struct
}

// validate and unmarshal the JSON
func retrieveJSON(path string) (map[string]map[string]map[string]any, error) {
	var bytes []byte
	data := make(map[string]map[string]map[string]any) // data container
	cleanFilePath, err := utils.CleanFilePath(path)    // clean the file path and check for errors
	if err != nil {                                    // error check
		return nil, err
	}
	if _, err = os.Stat(cleanFilePath); os.IsNotExist(err) { // check if file exists
		return nil, err
	}
	bytes, err = os.ReadFile(cleanFilePath) // read the bytes from the file
	if err != nil {                         // error check
		return nil, err
	}
	err = json.Unmarshal(bytes, &data) // unmarshall bytes into container
	if err != nil {                    // error check
		return nil, err
	}
	return data, nil // return the data for validation
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
	var err error
	if typeMap == nil || len(typeMap) != len(schema) {

		if typeMap == nil {
			err = errors.New("type map is nil")
		} else {
			err = SchemaError{expectedSchema: schema, actualSchema: typeMap}
		}
		return err
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
