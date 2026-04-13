package strategy

import (
	"fmt"
	"github.com/Keith1039/dbvg/utils"
	randomDataTime "github.com/duktig-solutions/go-random-date-generator"
	"github.com/golang-module/carbon"
	regen "github.com/zach-klippenstein/goregen"
	"math/rand/v2"
	"regexp"
	"strings"
	"time"
)

// RequiredStrategy is a type of strategy that requires a non-nil value to be given to it as input
type RequiredStrategy struct {
	*defaultStrategy
}

// ExecuteStrategy for the RequiredStrategy executes and returns it's strategy with its value field as input.
// Before the strategy is run, the program checks the validity using CheckCriteria.
// If an error occurs at any point, the program logs the failure and exits using log.Fatal
func (s *RequiredStrategy) ExecuteStrategy() (any, error) {
	if s.Value == nil {
		return nil, RequiredValueNilError{}
	}
	return s.defaultStrategy.ExecuteStrategy()
}

func (s *RequiredStrategy) CheckCriteria() error {
	if s.Value == nil {
		return RequiredValueNilError{}
	}
	return s.defaultStrategy.CheckCriteria()
}

func staticStrategy(val any) (any, error) {
	return val, nil
}

func staticCriteria[T any](val any) error {
	switch v := val.(type) {
	case T:
		return nil
	default:
		return UnexpectedTypeError{ExpectedType: fmt.Sprintf("%T", *new(T)), ActualType: fmt.Sprintf("%T", v)}
	}
}

func randomFloatCriteria(val any) error {
	switch t := val.(type) {
	case []float64:
		if len(t) == 2 {
			if t[0] > t[1] { // check to see if the bound works
				return RandomBoundError{LowerBound: t[0], UpperBound: t[1]}
			}
			return nil
		} else {
			return UnexpectedArrayLengthError{ExpectedLength: 2, ActualLength: len(t)}
		}
	default:
		return UnexpectedTypeError{ExpectedType: "[]float", ActualType: fmt.Sprintf("%T", t)}
	}
}

func randomIntCriteria(val any) error {
	switch t := val.(type) {
	case []int:
		if len(t) == 2 {
			if t[0] > t[1] { // check to see if the bound works
				return RandomBoundError{LowerBound: t[0], UpperBound: t[1]}
			}
			return nil
		} else {
			return UnexpectedArrayLengthError{ExpectedLength: 2, ActualLength: len(t)}
		}
	default:
		return UnexpectedTypeError{ExpectedType: "[]int", ActualType: fmt.Sprintf("%T", t)}
	}
}

// INT Type

// NewStaticIntStrategy defines and returns a Strategy of type RequiredStrategy meant to handle code "STATIC" for type "INT"
func NewStaticIntStrategy() ValueStrategy {
	return &RequiredStrategy{defaultStrategy: &defaultStrategy{Strategy: staticStrategy, Criteria: staticCriteria[int]}}
}

func randomIntStrategy(val any) (any, error) {
	intArr := val.([]int)
	return intArr[0] + rand.IntN(intArr[1]-intArr[0]), nil
}

// NewRandomIntStrategy defines and returns a Strategy of type RequiredStrategy meant to handle code "RANDOM" for type "INT"
func NewRandomIntStrategy() ValueStrategy {
	return &RequiredStrategy{defaultStrategy: &defaultStrategy{Strategy: randomIntStrategy, Criteria: randomIntCriteria}}
}

//

// Float

// NewStaticFloatStrategy defines and returns a Strategy of type RequiredStrategy meant to handle code "STATIC" for type "FLOAT"
func NewStaticFloatStrategy() ValueStrategy {
	return &RequiredStrategy{defaultStrategy: &defaultStrategy{Strategy: staticStrategy, Criteria: staticCriteria[float64]}}
}

// NewRandomFloatStrategy defines and returns a Strategy of type RequiredStrategy meant to handle code "RANDOM" for type "FLOAT"
func randomFloatStrategy(val any) (any, error) {
	floatArr := val.([]float64)
	return floatArr[0] + rand.Float64()*(floatArr[1]-floatArr[0]), nil
}

//

func NewRandomFloatStrategy() ValueStrategy {
	return &RequiredStrategy{defaultStrategy: &defaultStrategy{Strategy: randomFloatStrategy, Criteria: randomFloatCriteria}}
}

//

// Bool Type

// NewStaticBoolStrategy defines and returns a Strategy of type RequiredStrategy meant to handle code "STATIC" for type "BOOL"
func NewStaticBoolStrategy() ValueStrategy {
	return &RequiredStrategy{defaultStrategy: &defaultStrategy{Strategy: staticStrategy, Criteria: staticCriteria[bool]}}
}

//

// Varchar Type

// NewStaticVarcharStrategy defines and returns a Strategy of type RequiredStrategy meant to handle code "STATIC" for type "VARCHAR"
func NewStaticVarcharStrategy() ValueStrategy {
	return &RequiredStrategy{defaultStrategy: &defaultStrategy{Strategy: staticStrategy, Criteria: staticCriteria[string]}}
}

func regexStrategy(val any) (any, error) {
	expression := val.(string)
	return regen.Generate(expression)
}

func regexCriteria(val any) error {
	switch t := val.(type) {
	case string:
		_, err := regexp.Compile(t)
		if err != nil {
			return InvalidRegexError{Regex: t, Err: err}
		} else {
			return nil
		}
	default:
		return UnexpectedTypeError{ExpectedType: "string", ActualType: fmt.Sprintf("%T", t)}
	}
}

// NewRegexStrategy defines and returns a Strategy of type RequiredStrategy meant to handle code "REGEX" for type "VARCHAR"
func NewRegexStrategy() ValueStrategy {
	return &RequiredStrategy{defaultStrategy: &defaultStrategy{Strategy: regexStrategy, Criteria: regexCriteria}}
}

//

// Date Type
func staticDateCriteria(val any) error {
	switch t := val.(type) {
	case string:
		c := carbon.Parse(t)
		if c.Error != nil {
			return ImproperDateStringFormatError{DateString: t}
		}
		return nil
	default:
		return UnexpectedTypeError{ExpectedType: "string", ActualType: utils.GetStringType(t)}
	}
}

func staticDateStrategy(val any) (any, error) {
	v, ok := val.(string)
	if !ok {
		return nil, UnexpectedTypeError{ExpectedType: "string", ActualType: v}
	}
	c := carbon.Parse(v)
	if c.Error != nil {
		return nil, ImproperDateStringFormatError{DateString: v}
	}
	return c.ToStdTime(), nil
}

// NewStaticDateStrategy defines and returns a Strategy of type RequiredStrategy meant to handle code "STATIC" for type "DATE"
func NewStaticDateStrategy() ValueStrategy {
	return &RequiredStrategy{defaultStrategy: &defaultStrategy{Strategy: staticDateStrategy, Criteria: staticDateCriteria}}
}

func randomDateCriteria(val any) error {
	switch t := val.(type) {
	case []string:
		if len(t) == 2 {
			t1, err := utils.GetTimeFromString(t[0])
			if err != nil {
				return ImproperDateStringFormatError{DateString: t[0]}
			}
			t2, err2 := utils.GetTimeFromString(t[1])
			if err2 != nil {
				return ImproperDateStringFormatError{DateString: t[1]}
			}
			if t1.After(t2) { // check to see if the bound works
				return RandomBoundError{LowerBound: t[0], UpperBound: t[1]}
			}
			return nil
		} else {
			return UnexpectedArrayLengthError{ExpectedLength: 2, ActualLength: len(t)}
		}
	default:
		return UnexpectedTypeError{ExpectedType: "[]string", ActualType: fmt.Sprintf("%T", t)}
	}
}

func randomDateStrategy(val any) (any, error) {
	dates := val.([]string)
	lowerBound, err := utils.GetTimeFromString(dates[0])
	if err != nil {
		return nil, err
	}
	lowerBoundStr := lowerBound.Format(time.DateOnly)
	upperBound, err := utils.GetTimeFromString(dates[1])
	if err != nil {
		return nil, err
	}
	upperBoundStr := upperBound.Format(time.DateOnly)
	date, err := randomDataTime.GenerateDate(lowerBoundStr, upperBoundStr)
	if err != nil {
		return nil, err
	}
	return carbon.Parse(date).ToStdTime(), nil
}

// NewRandomDateStrategy defines and returns a Strategy of type RequiredStrategy meant to handle code "RANDOM" for type "DATE"
func NewRandomDateStrategy() ValueStrategy {
	return &RequiredStrategy{defaultStrategy: &defaultStrategy{Strategy: randomDateStrategy, Criteria: randomDateCriteria}}
}

// Time type

func staticTimeCriteria(val any) error {
	switch t := val.(type) {
	case string:
		t = strings.TrimSpace(t)
		_, err := time.Parse(time.TimeOnly, t)
		if err != nil {
			return ImproperTimeStringFormatError{TimeString: t}
		}
		return nil
	default:
		return UnexpectedTypeError{ExpectedType: "string", ActualType: utils.GetStringType(t)}
	}
}

func staticTimeStrategy(val any) (any, error) {
	timeString := val.(string)
	t, err := time.Parse(time.TimeOnly, timeString)
	if err != nil {
		return nil, ImproperTimeStringFormatError{TimeString: timeString}
	}
	return t, nil
}

// NewStaticTimeStrategy defines and returns a Strategy of type RequiredStrategy meant to handle code "STATIC" for type "TIME"
func NewStaticTimeStrategy() ValueStrategy {
	return &RequiredStrategy{defaultStrategy: &defaultStrategy{Strategy: staticTimeStrategy, Criteria: staticTimeCriteria}}
}

func randomTimeCriteria(val any) error {
	switch t := val.(type) {
	case []string:
		if len(t) == 2 {
			t[0] = strings.TrimSpace(t[0])
			t[1] = strings.TrimSpace(t[1])
			t1, err := time.Parse(time.TimeOnly, t[0])
			if err != nil {
				return ImproperTimeStringFormatError{TimeString: t[0]}
			}
			t2, err := time.Parse(time.TimeOnly, t[1])
			if err != nil {
				return ImproperTimeStringFormatError{TimeString: t[1]}
			}
			if t1.After(t2) { // check to see if the bound works
				return RandomBoundError{LowerBound: t[0], UpperBound: t[1]}
			}
			return nil
		} else {
			return UnexpectedArrayLengthError{ExpectedLength: 2, ActualLength: len(t)}
		}
	default:
		return UnexpectedTypeError{ExpectedType: "[]string", ActualType: fmt.Sprintf("%T", t)}
	}
}

func randomTimeStrategy(val any) (any, error) {
	t1, err := time.Parse(time.TimeOnly, "12:00:00")
	if err != nil {
		return nil, err
	}
	t2, err := time.Parse(time.TimeOnly, "23:00:00")
	if err != nil {
		return nil, err
	}
	duration := t2.Sub(t1)
	r := int64(duration / time.Second) // the range between the 2 times as an int
	num := rand.N(r)
	t3 := t1.Add(time.Duration(num) * time.Second)
	return t3, nil
}

// NewRandomTimeStrategy defines and returns a Strategy of type RequiredStrategy meant to handle code "RANDOM" for type "TIME"
func NewRandomTimeStrategy() ValueStrategy {
	return &RequiredStrategy{defaultStrategy: &defaultStrategy{Strategy: randomTimeStrategy, Criteria: randomTimeCriteria}}
}
