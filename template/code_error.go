package template

import (
	"fmt"
)

// RandomBoundError is an error given when the lower bound of a "RANDOM" code is greater than the upper bound
type RandomBoundError struct {
	lowerBound any
	upperBound any
}

func (e RandomBoundError) Error() string {
	return fmt.Sprintf("lower bound of '%v' is greater than upper bound of '%v'", e.lowerBound, e.upperBound)
}

// UnexpectedArrayLengthError is an error given when the length of the received array doesn't match the expected length
type UnexpectedArrayLengthError struct {
	expectedLength int
	actualLength   int
}

func (e UnexpectedArrayLengthError) Error() string {
	return fmt.Sprintf("expected array of length %d, received an array of length %d", e.expectedLength, e.actualLength)
}

// InvalidRegexError is an error given when the regex string, when given to `regexp.compile()`, returns an error
type InvalidRegexError struct {
	regexStr string
}

func (e InvalidRegexError) Error() string {
	return fmt.Sprintf("string '%s' is not proper regex", e.regexStr)
}

// ImproperDateStringFormatError is an error given when the given date string doesn't match the RFC3339
type ImproperDateStringFormatError struct {
	dateStr string
}

func (e ImproperDateStringFormatError) Error() string {
	return fmt.Sprintf("date string '%s' does not follow RFC3339 convention", e.dateStr)
}
