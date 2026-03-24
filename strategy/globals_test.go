package strategy_test

import (
	"errors"
	"fmt"
	"github.com/Keith1039/dbvg/strategy"
	"testing"
)

type factories interface {
	func() strategy.Strategy | func() strategy.ValueStrategy
}

func attemptDeletionForMap[T factories](m map[string]map[string]T) error {
	var err error
	for columnType, values := range m {
		for code, _ := range values {
			err = strategy.DeleteStrategy(columnType, code)
			if err == nil {
				return errors.New(fmt.Sprintf("sucessfully deleated a default, columnType: '%s', code: '%s'", columnType, code))
			}
		}
	}
	return nil
}
func TestDeleteStrategy(t *testing.T) {
	// test to see if defaults are up to date
	override := strategy.GetOverrideCodeMap()
	t.Logf("Beginning deletions on override code map: %v", override)
	err := attemptDeletionForMap(override)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("All deletions on override code map failed as expected")

	optional := strategy.GetOptionalCodeMap()
	t.Logf("Beginning tests on optional code map: %v", optional)
	err = attemptDeletionForMap(optional)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("All deletions on optional code map failed as expected")

	required := strategy.GetRequiredCodeMap()
	t.Logf("Beginning tests on required code map: %v", required)
	err = attemptDeletionForMap(required)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("All deletions on required code map failed as expected")

	// check if no op holds
	err = strategy.DeleteStrategy("giberrish", "blah")
	if err != nil {
		t.Fatal("if the strategy cannot be found, it's a no opp and nil should be returned")
	}
}

func TestGetStrategy(t *testing.T) {
	s, err := strategy.GetStrategy("", "")
	if err == nil {
		t.Fatalf("expected error for unsupported column type")
	}
	s, err = strategy.GetStrategy("int", "asddfaafaffd")
	if !errors.As(err, &strategy.UnsupportedCodeError{}) {
		t.Fatalf("expected 'UnsupportedCodeError' but received %v", err)
	}
	s, err = strategy.GetStrategy("          int        ", " random       ")
	if err != nil {
		t.Fatalf("expected no error but received %v", err)
	}
	s, err = strategy.GetStrategy("int", "null")
	if err != nil {
		t.Fatalf("expected no error but received %v", err)
	}
	genericStrat := s.(*strategy.OverrideStrategy).Strategy
	s, err = strategy.GetStrategy("uuid", "null")
	if err != nil {
		t.Fatalf("expected no error but received %v", err)
	}
	uuidSpecificStrat := s.(*strategy.OverrideStrategy).Strategy
	val1, _ := genericStrat()
	val2, _ := uuidSpecificStrat()
	if val1 == val2 {
		t.Fatal("generic Null strategy should not return the same as uuid's Null strategy")
	}
}

func testWorkingStrategy() strategy.Strategy {
	return &strategy.OverrideStrategy{}
}

func testInvalidStrategy() strategy.Strategy {
	return &strategy.OptionalStrategy{}
}
func TestAddOverrideStrategy(t *testing.T) {
	err := strategy.AddNewOverrideStrategy("something", "new", nil)
	if err == nil {
		t.Fatalf("expected error for unsupported column type")
	}
	err = strategy.AddNewOverrideStrategy("int", "new", nil)
	if !errors.As(err, &strategy.RequiredParameterNilError{}) {
		t.Fatalf("expected error for required parameter nil, received %v", err)
	}
	err = strategy.AddNewOverrideStrategy("int", "new", testInvalidStrategy)
	if !errors.As(err, &strategy.ValueStrategyImplementedError{}) {
		t.Fatalf("expected error for ValueStrategyImplementedError, received %v", err)
	}
	err = strategy.AddNewOverrideStrategy("int", "null", testWorkingStrategy)
	if !errors.As(err, &strategy.ExistingStrategyError{}) {
		t.Fatalf("expected error for ExistingStrategyError, received %v", err)
	}
	err = strategy.AddNewOverrideStrategy("uuid", "null", testWorkingStrategy)
	if !errors.As(err, &strategy.ExistingStrategyError{}) {
		t.Fatalf("expected error for ExistingStrategyError, received %v", err)
	}
	err = strategy.AddNewOverrideStrategy("bool", "random", testWorkingStrategy)
	if !errors.As(err, &strategy.ExistingStrategyError{}) {
		t.Fatalf("expected error for ExistingStrategyError, received %v", err)
	}
	err = strategy.AddNewOverrideStrategy("int", "new", testWorkingStrategy)
	if err != nil {
		t.Fatal(err)
	}
	_, err = strategy.GetStrategy("int", "new")
	if err != nil {
		t.Fatalf("new code failed to be added with error [%v]", err)
	}
	err = strategy.DeleteStrategy("int", "new")
	if err != nil {
		t.Fatalf("delete strategy failed to be removed with error [%v]", err)
	}
}

func workingValueStrategy() strategy.ValueStrategy {
	return &strategy.OptionalStrategy{}
}
func TestAddNewOptionalStrategy(t *testing.T) {
	err := strategy.AddNewOptionalStrategy("something", "new", nil)
	if err == nil {
		t.Fatalf("expected error for unsupported column type")
	}
	err = strategy.AddNewOptionalStrategy("int", "new", nil)
	if !errors.As(err, &strategy.RequiredParameterNilError{}) {
		t.Fatalf("expected error for required parameter nil, received %v", err)
	}
	err = strategy.AddNewOptionalStrategy("int", "null", workingValueStrategy)
	if !errors.As(err, &strategy.ExistingStrategyError{}) {
		t.Fatalf("expected error for ExistingStrategyError, received %v", err)
	}
	err = strategy.AddNewOptionalStrategy("uuid", "null", workingValueStrategy)
	if !errors.As(err, &strategy.ExistingStrategyError{}) {
		t.Fatalf("expected error for ExistingStrategyError, received %v", err)
	}
	err = strategy.AddNewOptionalStrategy("int", "serial", workingValueStrategy)
	if !errors.As(err, &strategy.ExistingStrategyError{}) {
		t.Fatalf("expected error for ExistingStrategyError, received %v", err)
	}
	err = strategy.AddNewOptionalStrategy("int", "something new", workingValueStrategy)
	if err != nil {
		t.Fatal(err)
	}

	_, err = strategy.GetStrategy("int", "something new")
	if err != nil {
		t.Fatalf("new code failed to be added with error [%v]", err)
	}
	err = strategy.DeleteStrategy("int", "something new")
	if err != nil {
		t.Fatalf("delete strategy failed to be removed with error [%v]", err)
	}
}

func workingRequiredValueStrategy() strategy.ValueStrategy {
	return &strategy.RequiredStrategy{}
}
func TestAddNewRequiredStrategy(t *testing.T) {
	err := strategy.AddNewRequiredStrategy("something", "new", nil)
	if err == nil {
		t.Fatalf("expected error for unsupported column type")
	}
	err = strategy.AddNewRequiredStrategy("int", "new", nil)
	if !errors.As(err, &strategy.RequiredParameterNilError{}) {
		t.Fatalf("expected error for required parameter nil, received %v", err)
	}
	err = strategy.AddNewRequiredStrategy("int", "null", workingRequiredValueStrategy)
	if !errors.As(err, &strategy.ExistingStrategyError{}) {
		t.Fatalf("expected error for ExistingStrategyError, received %v", err)
	}
	err = strategy.AddNewRequiredStrategy("uuid", "null", workingRequiredValueStrategy)
	if !errors.As(err, &strategy.ExistingStrategyError{}) {
		t.Fatalf("expected error for ExistingStrategyError, received %v", err)
	}
	err = strategy.AddNewRequiredStrategy("int", "random", workingRequiredValueStrategy)
	if !errors.As(err, &strategy.ExistingStrategyError{}) {
		t.Fatalf("expected error for ExistingStrategyError, received %v", err)
	}
	err = strategy.AddNewRequiredStrategy("varchar", "new", workingRequiredValueStrategy)
	if err != nil {
		t.Fatal(err)
	}
	_, err = strategy.GetStrategy("varchar", "new")
	if err != nil {
		t.Fatalf("new code failed to be added with error [%v]", err)
	}
	err = strategy.DeleteStrategy("varchar", "new")
	if err != nil {
		t.Fatalf("delete strategy failed to be removed with error [%v]", err)
	}

}

func TestDuplicateStrategyInDifferentMap(t *testing.T) {
	err := strategy.AddNewOptionalStrategy("Bool", "TEST", workingValueStrategy)
	if err != nil {
		t.Fatal(err)
	}
	err = strategy.AddNewRequiredStrategy("Bool", "TEST", workingRequiredValueStrategy)
	if !errors.As(err, &strategy.ExistingStrategyError{}) {
		t.Fatalf("expected error for ExistingStrategyError, received %v", err)
	}
	err = strategy.DeleteStrategy("bool", "TEST")
	if err != nil {
		t.Fatalf("delete strategy failed to be removed with error [%v]", err)
	}
}
