package strategy

import (
	"errors"
	"fmt"
	"github.com/Keith1039/dbvg/utils"
	"testing"
)

func intSerialTestRunner() *testRunner {
	t := testRunner{
		isOptional:     true,
		testValues:     []any{"", 20},
		expectedErrors: []error{UnexpectedTypeError{}},
	}
	t.evalCriteria = func(val any) error {
		val, ok := val.(int)
		if !ok {
			return UnexpectedTypeError{ExpectedType: "int", ActualType: fmt.Sprintf("%T", val)}
		}
		s, ok := t.strategy.(*SerialOptionalStrategy)
		if !ok {
			return errors.New("not serial")
		}
		if val == 20 && s.value.(int) == 21 {
			return nil
		} else {
			return errors.New("not serial2")
		}
	}
	return &t
}

var optionalRunnerMap = map[string]map[string]*testRunner{
	"INT": {"SERIAL": intSerialTestRunner()},
}

type testRunner struct {
	t              *testing.T
	strategy       ValueStrategy
	isOptional     bool
	testValues     []any               // testValues[0...n-2] = invalid values, testValues[n-1] is valid value
	expectedErrors []error             // an array of the errors we expect to run into
	evalCriteria   func(val any) error // function that takes in the value given by strategy and sees if it's valid
}

func (r *testRunner) Run() error {
	var evalVal any
	var err error
	err = r.handleOptional()
	if err != nil {
		return err
	}
	for i, val := range r.testValues {
		r.strategy.SetValue(val)
		err = r.strategy.CheckCriteria()
		if i == len(r.testValues)-1 {
			r.t.Logf("testing valid value %v", val)
			if err != nil {
				return err
			}
			evalVal, err = r.strategy.ExecuteStrategy()
			if err != nil {
				return err
			}
			err = r.evalCriteria(evalVal)
			if err != nil {
				return err
			}
		} else {
			r.t.Logf("testing invalid value %v", val)
			expectedErr := r.expectedErrors[i]
			if err == nil || !(utils.GetStringType(err) == utils.GetStringType(expectedErr)) {
				return UnexpectedTypeError{ExpectedType: utils.GetStringType(expectedErr), ActualType: utils.GetStringType(err)}
			}
		}
	}
	return nil
}

func (r *testRunner) handleOptional() error {
	var err error
	if r.isOptional {
		r.t.Log("beginning nil check for OptionalStrategy")
		// if it's optional it should handle nil
		r.strategy.SetValue(nil)
		_, err = r.strategy.ExecuteStrategy()
		if err != nil {
			r.t.Log("failed to handle 'nil' as value ")
			return err
		}
		r.t.Log("ending nil check for OptionalStrategy")
	}
	return nil
}

func TestOptionalStrategies(t *testing.T) {
	for colType, codeMap := range optionalRunnerMap {
		t.Log(".........................................")
		t.Logf("Beginning test suite for type '%s'...", colType)
		for code, runner := range codeMap {
			runner.t = t
			t.Logf("testing code '%s'...", code)
			s, err := GetStrategy(colType, code)
			if err != nil {
				t.Fatal(err)
			}
			sVal, ok := s.(ValueStrategy)
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

func TestEnforceNonDuplicates(t *testing.T) {
	s, err := GetStrategy("int", "serial")
	if err != nil {
		t.Fatal(err)
	}
	s2, err := GetStrategy("int", "serial")
	if err != nil {
		t.Fatal(err)
	}
	if s == s2 {
		t.Fatal("2 int Serial strategies cannot point to the same underlying strategy")
	}

}
