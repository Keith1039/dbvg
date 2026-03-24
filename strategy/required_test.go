package strategy_test

import (
	"errors"
	"fmt"
	"github.com/Keith1039/dbvg/strategy"
	"github.com/Keith1039/dbvg/utils"
	"github.com/golang-module/carbon"
	"regexp"
	"testing"
)

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
		arr := s.Value.([]string)
		t1, _ := utils.GetTimeFromString(arr[0])
		t2, _ := utils.GetTimeFromString(arr[1])
		t3, err := utils.GetTimeFromString(val.(string))
		if err != nil {
			return err
		}
		if t3.Before(t1) || t3.After(t2) {
			return errors.New(fmt.Sprintf("generated value '%v' is out of bounds of value '%v", t3, arr))
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
		return staticEvaluator(val, s.Value)
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
		lowerBound := s.Value.([]int)[0]
		upperBound := s.Value.([]int)[1]
		intVal, ok := val.(int)
		if !ok {
			return errors.New("not int")
		}
		if intVal < lowerBound || intVal > upperBound {
			return errors.New(fmt.Sprintf("generated value '%v' is out of bounds of value '%v", val, s.Value))
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
		lowerBound := s.Value.([]float64)[0]
		upperBound := s.Value.([]float64)[1]
		floatVal, ok := val.(float64)
		if !ok {
			return errors.New("not float")
		}
		if floatVal < lowerBound || floatVal > upperBound {
			return errors.New(fmt.Sprintf("generated value '%v' is out of bounds of value '%v", val, s.Value))
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
		regex := s.Value.(string)
		regexStr := val.(string)
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

var requiredRunnerMap = map[string]map[string]*testRunner{
	"BOOL": {
		"STATIC": boolStaticTestRunner(),
	},
	"DATE": {
		"RANDOM": dateRandomTestRunner(),
		"STATIC": dateStaticTestRunner(),
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
