package strategy

import (
	"fmt"
	"strings"
)

// UnexpectedTypeError is an error given when the given type doesn't match an expected type
type UnexpectedTypeError struct {
	ExpectedType string
	ActualType   string
}

func (err UnexpectedTypeError) Error() string {
	return fmt.Sprintf("expected type %s but received type %s", err.ExpectedType, err.ActualType)
}

// UnsupportedCodeError is an error given when the code given by the template isn't supported by its type
type UnsupportedCodeError struct {
	Code       string
	ColumnType string
}

func (e UnsupportedCodeError) Error() string {
	return fmt.Sprintf("code '%s' is not supported for the type '%s'", e.Code, e.ColumnType)
}

// UnsupportedTypeError  is an error given when the value given by the template doesn't match a supported type for that code
type UnsupportedTypeError struct {
	Type           string
	Code           string
	SupportedTypes []string
}

func (e UnsupportedTypeError) Error() string {
	return fmt.Sprintf("type '%s' is not supported for the code '%s'. Types supported are [%s]", e.Type, e.Code, strings.Join(e.SupportedTypes, ","))
}

// RandomBoundError is an error given when the lower bound of a "RANDOM" code is greater than the upper bound
type RandomBoundError struct {
	LowerBound any
	UpperBound any
}

func (e RandomBoundError) Error() string {
	return fmt.Sprintf("lower bound of '%v' is greater than upper bound of '%v'", e.LowerBound, e.UpperBound)
}

// UnexpectedArrayLengthError is an error given when the length of the received array doesn't match the expected length
type UnexpectedArrayLengthError struct {
	ExpectedLength int
	ActualLength   int
}

func (e UnexpectedArrayLengthError) Error() string {
	return fmt.Sprintf("expected array of length %d, received an array of length %d", e.ExpectedLength, e.ActualLength)
}

// InvalidRegexError is an error given when the regex string, when given to `regexp.compile()`, returns an error
type InvalidRegexError struct {
	Regex string
	Err   error
}

func (e InvalidRegexError) Error() string {
	return fmt.Sprintf("string '%s' is not proper regex, failed with %v", e.Regex, e.Err)
}

// ImproperDateStringFormatError is an error given when the given date string doesn't match any known format
type ImproperDateStringFormatError struct {
	DateString string
}

func (e ImproperDateStringFormatError) Error() string {
	return fmt.Sprintf("date string '%s' cannot be parsed via carbon", e.DateString)
}

// ImproperTimeStringFormatError is an error given when the given time string doesn't match the time.TimeOnly layout
type ImproperTimeStringFormatError struct {
	TimeString string
}

func (e ImproperTimeStringFormatError) Error() string {
	return fmt.Sprintf("time string '%s' does not fit the layout of 'HH:MM:SS' or was beyond the acceptable range of '00:00:00' - '23:59:59'", e.TimeString)
}
