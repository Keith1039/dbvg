package strategy

import "fmt"

// UnspecifiedStrategyError is an error given when a Strategy cannot be executed since it's defined Strategy parameter is null
type UnspecifiedStrategyError struct{}

func (e UnspecifiedStrategyError) Error() string {
	return "parameter 'Strategy' cannot be executed as it is undefined"
}

type UnspecifiedCriteriaError struct{}

// UnspecifiedCriteriaError is an error given when a Strategy cannot execute its criteria since it's left undefined
func (e UnspecifiedCriteriaError) Error() string {
	return "parameter 'Criteria' cannot be evaluated as it is undefined"
}

// RequiredValueNilError is an error given when a nil value is given to a RequiredStrategy and isn't filtered out by its criteria
type RequiredValueNilError struct{}

func (e RequiredValueNilError) Error() string {
	return "value cannot be nil for type 'RequiredStrategy'"
}

// ExistingStrategyError is an error given when a user tries to overwrite an existing strategy on the internal maps
type ExistingStrategyError struct {
	ColumnType string
	Code       string
}

func (e ExistingStrategyError) Error() string {
	return fmt.Sprintf("code '%s' already exists for column type '%s'", e.Code, e.ColumnType)
}

// ValueStrategyImplementedError is an error given when the user tries to add a ValueStrategy to the internal override map
type ValueStrategyImplementedError struct {
}

func (e ValueStrategyImplementedError) Error() string {
	return "strategy implements 'ValueStrategy' and thus cannot be added as override strategy"
}

// RequiredParameterNilError is an error given when a required parameter is set to nil
type RequiredParameterNilError struct {
	Name string
}

func (e RequiredParameterNilError) Error() string {
	return fmt.Sprintf("required parameter '%s' cannot be nil", e.Name)
}

type DeleteDefaultError struct {
	columnType string
	code       string
}

func (e DeleteDefaultError) Error() string {
	return fmt.Sprintf("cannot delete the following strategy for column type: '%s' and code: '%s' as it is a default code", e.columnType, e.code)
}

// NotInRangeError is given when a value isn't in a predefined range which represented as a string
type NotInRangeError struct {
	value    any
	rangeStr string
}

func (e NotInRangeError) Error() string {
	return fmt.Sprintf("value '%v' is not within the accepted range: '%s'", e.value, e.rangeStr)
}

func wrapError(columnType string, code string, err error) error {
	return fmt.Errorf("for column type '%s' and code '%s': [%w]", columnType, code, err)
}
