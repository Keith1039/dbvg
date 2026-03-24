package strategy

import (
	"fmt"
)

// OptionalStrategy is a type of Strategy that can take in either nil or a valid input as their value
type OptionalStrategy struct {
	*defaultStrategy
}

type SerialOptionalStrategy struct {
	*defaultStrategy
}

func (s *SerialOptionalStrategy) ExecuteStrategy() (any, error) {
	cur := s.Value.(int)
	s.Value = s.Value.(int) + 1
	return cur, nil
}

func (s *SerialOptionalStrategy) SetValue(val any) {
	// assume criteria works
	if val == nil {
		s.Value = 0
	} else {
		s.Value = val
	}
}

func serialIntCriteria(val any) error {
	switch t := val.(type) {
	case int:
		return nil
	default:
		return UnexpectedTypeError{ExpectedType: "int", ActualType: fmt.Sprintf("%T", t)}
	}
}

// NewSerialStrategy defines and returns a ValueStrategy of type SerialOptionalStrategy to handle the "SERIAL" code for the "INT" type
func NewSerialStrategy() ValueStrategy {
	return &SerialOptionalStrategy{defaultStrategy: &defaultStrategy{Criteria: serialIntCriteria}}
}
