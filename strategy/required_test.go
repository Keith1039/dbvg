package strategy_test

import (
	"errors"
	"fmt"
	"github.com/Keith1039/dbvg/strategy"
	"github.com/Keith1039/dbvg/utils"
	"github.com/golang-module/carbon"
	"log"
	"regexp"
	"testing"
	"time"
)

func init() {
	// make sure a test runner exists for each Strategy
	req := strategy.GetRequiredCodeMap()
	for typeString, v := range req {
		for code := range v {
			if _, ok := requiredRunnerMap[typeString][code]; !ok {
				log.Fatalf("missing tests for required strategy: code '%s' of type '%s'", code, typeString)
			}
		}
	}
}

var requiredRunnerMap = map[string]map[string]*testRunner{
	"BOOL": {
		"STATIC": boolStaticTestRunner(),
	},
	"DATE": {
		"RANDOM": dateRandomTestRunner(),
		"STATIC": dateStaticTestRunner(),
	},
	"TIME": {
		"RANDOM": timeRandomTestRunner(),
		"STATIC": timeStaticTestRunner(),
	},
	"INT": {
		"RANDOM": intRandomTestRunner(),
		"STATIC": intStaticTestRunner(),
	},
	"FLOAT": {
		"RANDOM": floatRandomTestRunner(),
		"STATIC": floatStaticTestRunner(),
	},
	"VARCHAR": {
		"STATIC": varcharStaticTestRunner(),
		"REGEX":  varcharRegexTestRunner(),
	},
}

func TestRequiredStrategy_Behavior(t *testing.T) {
	// test that a value check is performed when attempting to execute strategy or check criteria
	s := strategy.RequiredStrategy{}
	err := s.CheckCriteria()
	if !errors.As(err, &strategy.RequiredValueNilError{}) {
		t.Fatalf("expected 'RequiredValueNilError', got %v", err)
	}
	_, err = s.ExecuteStrategy()
	if !errors.As(err, &strategy.RequiredValueNilError{}) {
		t.Fatalf("expected 'RequiredValueNilError', got %v", err)
	}
}

func TestRequiredStrategies(t *testing.T) {
	for colType, codeMap := range requiredRunnerMap {
		t.Log(".........................................")
		t.Logf("Beginning test suite for type '%s'...", colType)
		for code, runner := range codeMap {
			runner.t = t
			t.Logf("testing code '%s'...", code)
			s, err := strategy.GetStrategy(colType, code)
			if err != nil {
				t.Fatal(err)
			}
			sVal, ok := s.(strategy.ValueStrategy)
			if !ok {
				t.Fatal("failed to assert value strategy")
			}
			runner.strategy = sVal
			err = runner.Run()
			if err != nil {
				t.Fatal(wrapError(colType, code, err))
			}
			t.Logf("tests for code '%s' ended successfully!", code)
		}
		t.Logf("Ending test suite for type '%s'...", colType)
		t.Log(".........................................\n")
	}
}

type expectedValueError struct {
	expectedValue any
	actualValue   any
}

func (e expectedValueError) Error() string {
	return fmt.Sprintf("expected value '%v' but got '%v'", e.expectedValue, e.actualValue)
}

func staticEvaluator(expectedVal any, actualVal any) error {
	if expectedVal != actualVal {
		return expectedValueError{expectedVal, actualVal}
	}
	return nil
}

func boolStaticTestRunner() *testRunner {
	t := testRunner{
		testValues:     []any{"", true},
		expectedErrors: []error{strategy.UnexpectedTypeError{}},
	}

	t.evalCriteria = func(val any) error {
		s, ok := t.strategy.(*strategy.RequiredStrategy)
		if !ok {
			return errors.New("strategy could not be cast to 'RequiredStrategy'")
		}
		return staticEvaluator(val, s.Value)
	}

	return &t
}

func dateRandomTestRunner() *testRunner {
	t := testRunner{
		testValues: []any{"", []string{"", "", ""}, []string{"03-01-2025", "2025-01-02"}, []string{"2025-04-06", "02-12-2025"},
			[]string{"2026-01-02", "2025-01-02"}, []string{"2025-01-10", "2026-01-02"}},
		expectedErrors: []error{strategy.UnexpectedTypeError{}, strategy.UnexpectedArrayLengthError{}, strategy.ImproperDateStringFormatError{}, strategy.ImproperDateStringFormatError{}, strategy.RandomBoundError{}},
	}
	t.evalCriteria = func(val any) error {
		s, ok := t.strategy.(*strategy.RequiredStrategy)
		if !ok {
			return errors.New("strategy could not be cast to 'RequiredStrategy'")
		}
		arr, ok := s.Value.([]string)
		if !ok {
			return strategy.UnexpectedTypeError{ExpectedType: "[]string", ActualType: utils.GetStringType(s.Value)}
		}
		t1, err := utils.GetTimeFromString(arr[0])
		if err != nil {
			return err
		}
		t2, err := utils.GetTimeFromString(arr[1])
		if err != nil {
			return err
		}
		t3, ok := val.(time.Time)
		if !ok {
			return strategy.UnexpectedTypeError{ExpectedType: "time.Time", ActualType: utils.GetStringType(val)}
		}
		if t3.Before(t1) || t3.After(t2) {
			return strategy.RandomBoundError{LowerBound: t1, UpperBound: t2}
		}
		return nil
	}
	return &t
}

func dateStaticTestRunner() *testRunner {
	t := testRunner{
		testValues:     []any{false, "01-02-2026", carbon.Parse("2026-03-01").ToRfc3339String()},
		expectedErrors: []error{strategy.UnexpectedTypeError{}, strategy.ImproperDateStringFormatError{}},
	}
	t.evalCriteria = func(val any) error {
		s, ok := t.strategy.(*strategy.RequiredStrategy)
		if !ok {
			return errors.New("strategy could not be cast to 'RequiredStrategy'")
		}
		date, ok := s.Value.(string)
		if !ok {
			return strategy.UnexpectedTypeError{ExpectedType: "string", ActualType: utils.GetStringType(s.Value)}
		}
		return staticEvaluator(val, carbon.Parse(date).ToStdTime())
	}
	return &t
}

func timeStaticTestRunner() *testRunner {
	t := testRunner{
		testValues:     []any{false, "24:00:00", "22:00:00"},
		expectedErrors: []error{strategy.UnexpectedTypeError{}, strategy.ImproperTimeStringFormatError{}},
	}

	t.evalCriteria = func(val any) error {
		s, ok := t.strategy.(*strategy.RequiredStrategy)
		if !ok {
			return errors.New("strategy could not be cast to 'RequiredStrategy'")
		}
		timeStr, ok := s.Value.(string)
		if !ok {
			return strategy.UnexpectedTypeError{ExpectedType: "string", ActualType: utils.GetStringType(s.Value)}
		}
		timeObj, err := time.Parse(time.TimeOnly, timeStr)
		if err != nil {
			return err
		}
		return staticEvaluator(val, timeObj)
	}

	return &t
}

func timeRandomTestRunner() *testRunner {
	t := testRunner{
		testValues: []any{"", []string{"", "", ""}, []string{"03-01-2025", "2025-01-02"},
			[]string{"00:00:00", "24:00:00"}, []string{"13:00:00", "00:00:00"},
			[]string{"00:00:00", "02:00:00"}},
		expectedErrors: []error{strategy.UnexpectedTypeError{}, strategy.UnexpectedArrayLengthError{}, strategy.ImproperTimeStringFormatError{},
			strategy.ImproperTimeStringFormatError{}, strategy.RandomBoundError{}},
	}
	t.evalCriteria = func(val any) error {
		s, ok := t.strategy.(*strategy.RequiredStrategy)
		if !ok {
			return errors.New("strategy could not be cast to 'RequiredStrategy'")
		}
		arr, ok := s.Value.([]string)
		if !ok {
			return strategy.UnexpectedTypeError{ExpectedType: "[]string", ActualType: utils.GetStringType(s.Value)}
		}
		t1, err := time.Parse(time.TimeOnly, arr[0])
		if err != nil {
			return err
		}
		t2, err := time.Parse(time.TimeOnly, arr[1])
		if err != nil {
			return err
		}
		t3, ok := val.(time.Time)
		if !ok {
			return strategy.UnexpectedTypeError{ExpectedType: "time.Time", ActualType: utils.GetStringType(val)}
		}
		if t3.Before(t1) || t3.After(t2) {
			return strategy.RandomBoundError{LowerBound: t1, UpperBound: t2}
		}
		return nil
	}
	return &t
}

func intRandomTestRunner() *testRunner {
	t := testRunner{
		testValues:     []any{"", []int{5, 10, 20}, []int{10, 5}, []int{5, 10}},
		expectedErrors: []error{strategy.UnexpectedTypeError{}, strategy.UnexpectedArrayLengthError{}, strategy.RandomBoundError{}},
	}
	t.evalCriteria = func(val any) error {
		s, ok := t.strategy.(*strategy.RequiredStrategy)
		if !ok {
			return errors.New("strategy could not be cast to 'RequiredStrategy'")
		}
		arr, ok := s.Value.([]int)
		if !ok {
			return strategy.UnexpectedTypeError{ExpectedType: "[]int", ActualType: utils.GetStringType(s.Value)}
		}
		lowerBound := arr[0]
		upperBound := arr[1]
		intVal, ok := val.(int)
		if !ok {
			return strategy.UnexpectedTypeError{ExpectedType: "int", ActualType: utils.GetStringType(val)}
		}
		if intVal < lowerBound || intVal > upperBound {
			return strategy.RandomBoundError{LowerBound: lowerBound, UpperBound: upperBound}
		}
		return nil
	}
	return &t
}

func intStaticTestRunner() *testRunner {
	t := testRunner{
		testValues:     []any{false, 20000},
		expectedErrors: []error{strategy.UnexpectedTypeError{}},
	}

	t.evalCriteria = func(val any) error {
		s, ok := t.strategy.(*strategy.RequiredStrategy)
		if !ok {
			return errors.New("strategy could not be cast to 'RequiredStrategy'")
		}
		return staticEvaluator(val, s.Value)
	}

	return &t
}

func floatRandomTestRunner() *testRunner {
	t := testRunner{
		testValues:     []any{"", []float64{5, 10.5, 20}, []float64{5.5, 5}, []float64{5.67, 10.9}},
		expectedErrors: []error{strategy.UnexpectedTypeError{}, strategy.UnexpectedArrayLengthError{}, strategy.RandomBoundError{}},
	}
	t.evalCriteria = func(val any) error {
		s, ok := t.strategy.(*strategy.RequiredStrategy)
		if !ok {
			return errors.New("strategy could not be cast to 'RequiredStrategy'")
		}
		arr, ok := s.Value.([]float64)
		if !ok {
			return strategy.UnexpectedTypeError{ExpectedType: "[]float64", ActualType: utils.GetStringType(s.Value)}
		}
		lowerBound := arr[0]
		upperBound := arr[1]
		floatVal, ok := val.(float64)
		if !ok {
			return strategy.UnexpectedTypeError{ExpectedType: "float64", ActualType: utils.GetStringType(val)}
		}
		if floatVal < lowerBound || floatVal > upperBound {
			return strategy.RandomBoundError{LowerBound: lowerBound, UpperBound: upperBound}
		}
		return nil
	}
	return &t
}

func floatStaticTestRunner() *testRunner {
	t := testRunner{
		testValues:     []any{false, 20.99},
		expectedErrors: []error{strategy.UnexpectedTypeError{}},
	}

	t.evalCriteria = func(val any) error {
		s, ok := t.strategy.(*strategy.RequiredStrategy)
		if !ok {
			return errors.New("strategy could not be cast to 'RequiredStrategy'")
		}
		return staticEvaluator(val, s.Value)
	}
	return &t
}

func varcharStaticTestRunner() *testRunner {
	t := testRunner{
		testValues:     []any{200, "something"},
		expectedErrors: []error{strategy.UnexpectedTypeError{}},
	}

	t.evalCriteria = func(val any) error {
		s, ok := t.strategy.(*strategy.RequiredStrategy)
		if !ok {
			return errors.New("strategy could not be cast to 'RequiredStrategy'")
		}
		return staticEvaluator(val, s.Value)
	}
	return &t
}

func varcharRegexTestRunner() *testRunner {
	t := testRunner{
		testValues:     []any{200, `([A-Z]{3}`, `[A-Z]{5}`},
		expectedErrors: []error{strategy.UnexpectedTypeError{}, strategy.InvalidRegexError{}},
	}

	t.evalCriteria = func(val any) error {
		s, ok := t.strategy.(*strategy.RequiredStrategy)
		if !ok {
			return errors.New("strategy could not be cast to 'RequiredStrategy'")
		}
		regex, ok := s.Value.(string)
		if !ok {
			return strategy.UnexpectedTypeError{ExpectedType: "string", ActualType: utils.GetStringType(s.Value)}
		}
		regexStr, ok := val.(string)
		if !ok {
			return strategy.UnexpectedTypeError{ExpectedType: "string", ActualType: utils.GetStringType(val)}
		}
		matched, err := regexp.MatchString(regex, regexStr)
		if err != nil {
			return err
		}
		if !matched {
			return errors.New("generated string failed to match original regex")
		}
		return nil
	}
	return &t
}
