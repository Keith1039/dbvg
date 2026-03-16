package template

import (
	"fmt"
	"math"
)

// preprocess takes in a pointer to any type and if that value is of []any, it tries to convert it to []int, []float or []string. If it fails
// it returns an error of type PreprocessError
func preprocess(val *any, columnType string) error {
	switch (*val).(type) {
	case float64:
		if columnType == "INT" {
			// TODO: should give warning if rounding changes value
			*val = int((*val).(float64)) // convert to int
		}
	case []interface{}:
		// check if this is string
		if (columnType == "VARCHAR" || columnType == "DATE") && checkForTypeHomogeny(*val, "string") { // check if it's a string arr
			convertToStringArray(val) // convert it to []string under the hood
		} else if columnType == "INT" && checkIfNumericArray(*val) { // check if it's a numeric array
			convertToIntArray(val) // convert it to []int under the hood
		} else if columnType == "FLOAT" && checkIfNumericArray(*val) {
			convertToFloatArray(val) // convert it to []float64 under the hood
		} else {
			return PreprocessError{val: *val} // return a preprocessing error
		}
	}
	return nil
}

// checkForTypeHomogeny checks if all elements in an []any array are of the same type
func checkForTypeHomogeny(val any, columnType string) bool {
	for _, arrVal := range val.([]any) {
		if fmt.Sprintf("%T", arrVal) != columnType {
			return false
		}
	}
	return true
}

// checkIfNumericArray calls checkForTypeHomogeny but for float64 as the column type
func checkIfNumericArray(val any) bool {
	return checkForTypeHomogeny(val, "float64")
}

// convertToStringArray converts the given []any array to []string
func convertToStringArray(val *any) {
	strArr := make([]string, len((*val).([]any)))
	for i, arrVal := range (*val).([]any) {
		strArr[i] = arrVal.(string)
	}
	*val = strArr
}

// convertToIntArray converts the given []any array to []int and logs warning when rounding occurs
func convertToIntArray(val *any) {
	intArr := make([]int, len((*val).([]any)))
	for i, arrVal := range (*val).([]any) {
		if math.Floor(arrVal.(float64)) != arrVal.(float64) { // if the value changes when floored, it wasn't an int
			// warn about rounding
		}
		intArr[i] = int(arrVal.(float64)) // set the value
	}
	*val = intArr
}

// convertToFloatArray converts the given []any array to []float
func convertToFloatArray(val *any) {
	floatArr := make([]float64, len((*val).([]any)))
	for i, arrVal := range (*val).([]any) {
		floatArr[i] = arrVal.(float64)
	}
	*val = floatArr
}
