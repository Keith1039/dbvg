package strategy

import (
	"errors"
	"testing"
)

func TestGetStrategy(t *testing.T) {
	s, err := GetStrategy("", "")
	if err == nil {
		t.Fatalf("expected error for unsupported column type")
	}
	s, err = GetStrategy("int", "asddfaafaffd")
	if !errors.As(err, &UnsupportedCodeError{}) {
		t.Fatalf("expected 'UnsupportedCodeError' but received %v", err)
	}
	s, err = GetStrategy("          int        ", " random       ")
	if err != nil {
		t.Fatalf("expected no error but received %v", err)
	}
	s, err = GetStrategy("int", "null")
	if err != nil {
		t.Fatalf("expected no error but received %v", err)
	}
	genericStrat := s.(*OverrideStrategy).Strategy
	s, err = GetStrategy("uuid", "null")
	if err != nil {
		t.Fatalf("expected no error but received %v", err)
	}
	uuidSpecificStrat := s.(*OverrideStrategy).Strategy
	val1, _ := genericStrat()
	val2, _ := uuidSpecificStrat()
	if val1 == val2 {
		t.Fatal("generic Null strategy should not return the same as uuid's Null strategy")
	}
}

func testWorkingStrategy() Strategy {
	return &OverrideStrategy{}
}

func testInvalidStrategy() Strategy {
	return &OptionalStrategy{}
}
func TestAddOverrideStrategy(t *testing.T) {
	err := AddNewOverrideStrategy("something", "new", nil)
	if err == nil {
		t.Fatalf("expected error for unsupported column type")
	}
	err = AddNewOverrideStrategy("int", "new", nil)
	if !errors.As(err, &RequiredParameterNilError{}) {
		t.Fatalf("expected error for required parameter nil, received %v", err)
	}
	err = AddNewOverrideStrategy("int", "new", testInvalidStrategy)
	if !errors.As(err, &ValueStrategyImplementedError{}) {
		t.Fatalf("expected error for ValueStrategyImplementedError, received %v", err)
	}
	err = AddNewOverrideStrategy("int", "null", testWorkingStrategy)
	if !errors.As(err, &ExistingStrategyError{}) {
		t.Fatalf("expected error for ExistingStrategyError, received %v", err)
	}
	err = AddNewOverrideStrategy("uuid", "null", testWorkingStrategy)
	if !errors.As(err, &ExistingStrategyError{}) {
		t.Fatalf("expected error for ExistingStrategyError, received %v", err)
	}
	err = AddNewOverrideStrategy("bool", "random", testWorkingStrategy)
	if !errors.As(err, &ExistingStrategyError{}) {
		t.Fatalf("expected error for ExistingStrategyError, received %v", err)
	}
	err = AddNewOverrideStrategy("int", "new", testWorkingStrategy)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := overrideCodeMap["INT"]["NEW"]; !ok {
		t.Fatal("new code failed to be added")
	}
	delete(overrideCodeMap["INT"], "NEW")
}

func workingValueStrategy() ValueStrategy {
	return &OptionalStrategy{}
}
func TestAddNewOptionalStrategy(t *testing.T) {
	err := AddNewOptionalStrategy("something", "new", nil)
	if err == nil {
		t.Fatalf("expected error for unsupported column type")
	}
	err = AddNewOptionalStrategy("int", "new", nil)
	if !errors.As(err, &RequiredParameterNilError{}) {
		t.Fatalf("expected error for required parameter nil, received %v", err)
	}
	err = AddNewOptionalStrategy("int", "null", workingValueStrategy)
	if !errors.As(err, &ExistingStrategyError{}) {
		t.Fatalf("expected error for ExistingStrategyError, received %v", err)
	}
	err = AddNewOptionalStrategy("uuid", "null", workingValueStrategy)
	if !errors.As(err, &ExistingStrategyError{}) {
		t.Fatalf("expected error for ExistingStrategyError, received %v", err)
	}
	err = AddNewOptionalStrategy("int", "serial", workingValueStrategy)
	if !errors.As(err, &ExistingStrategyError{}) {
		t.Fatalf("expected error for ExistingStrategyError, received %v", err)
	}
	err = AddNewOptionalStrategy("int", "something new", workingValueStrategy)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := optionalCodeMap["INT"]["SOMETHING NEW"]; !ok {
		t.Fatal("new code failed to be added")
	}
	delete(optionalCodeMap["INT"], "SOMETHING NEW")
}

func workingRequiredValueStrategy() ValueStrategy {
	return &RequiredStrategy{}
}
func TestAddNewRequiredStrategy(t *testing.T) {
	err := AddNewRequiredStrategy("something", "new", nil)
	if err == nil {
		t.Fatalf("expected error for unsupported column type")
	}
	err = AddNewRequiredStrategy("int", "new", nil)
	if !errors.As(err, &RequiredParameterNilError{}) {
		t.Fatalf("expected error for required parameter nil, received %v", err)
	}
	err = AddNewRequiredStrategy("int", "null", workingRequiredValueStrategy)
	if !errors.As(err, &ExistingStrategyError{}) {
		t.Fatalf("expected error for ExistingStrategyError, received %v", err)
	}
	err = AddNewRequiredStrategy("uuid", "null", workingRequiredValueStrategy)
	if !errors.As(err, &ExistingStrategyError{}) {
		t.Fatalf("expected error for ExistingStrategyError, received %v", err)
	}
	err = AddNewRequiredStrategy("int", "random", workingRequiredValueStrategy)
	if !errors.As(err, &ExistingStrategyError{}) {
		t.Fatalf("expected error for ExistingStrategyError, received %v", err)
	}
	err = AddNewRequiredStrategy("varchar", "new", workingRequiredValueStrategy)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := requiredCodeMap["VARCHAR"]["NEW"]; !ok {
		t.Fatal("new code failed to be added")
	}
	defer func() {
		delete(requiredCodeMap["VARCHAR"], "NEW")
	}()
}

func TestDuplicateStrategyInDifferentMap(t *testing.T) {
	err := AddNewOptionalStrategy("Bool", "TEST", workingValueStrategy)
	if err != nil {
		t.Fatal(err)
	}
	err = AddNewRequiredStrategy("Bool", "TEST", workingRequiredValueStrategy)
	if !errors.As(err, &ExistingStrategyError{}) {
		t.Fatalf("expected error for ExistingStrategyError, received %v", err)
	}
	delete(optionalCodeMap["BOOL"], "TEST")
}
