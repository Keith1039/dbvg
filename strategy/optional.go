package strategy

import (
	"fmt"
	"github.com/Keith1039/dbvg/utils"
)

// OptionalStrategy is a type of Strategy that can take in either nil or a valid input as their value
type OptionalStrategy struct {
	Default any
	*defaultStrategy
}

func (s *OptionalStrategy) SetValue(val any) {
	if val == nil {
		s.Value = s.Default
	} else {
		s.Value = val
	}
}

func serialIntCriteria(val any) error {
	switch t := val.(type) {
	case int:
		if t < 1 {
			return NotInRangeError{value: t, rangeStr: ">=1"}
		}
		return nil
	default:
		return UnexpectedTypeError{ExpectedType: "int", ActualType: fmt.Sprintf("%T", t)}
	}
}

// NewSerialStrategy defines and returns a ValueStrategy of type SerialOptionalStrategy to handle the "SERIAL" code for the "INT" type
func NewSerialStrategy() ValueStrategy {
	s := &OptionalStrategy{Default: 1, defaultStrategy: &defaultStrategy{Criteria: serialIntCriteria}}
	s.Strategy = func(val any) (any, error) {
		intVal, ok := val.(int)
		if !ok {
			return nil, UnexpectedTypeError{ExpectedType: "int", ActualType: utils.GetStringType(val)}
		}
		s.Value = intVal + 1
		return intVal, nil
	}
	return s
}
