package strategy_test

import (
	"errors"
	"fmt"
	"github.com/Keith1039/dbvg/strategy"
	"github.com/Keith1039/dbvg/utils"
	"log"
	"testing"
)

func implementsValueStrategy(s strategy.Strategy) bool {
	_, ok := s.(strategy.ValueStrategy)
	return ok
}

func init() {
	// make sure a test runner exists for each Strategy
	override := strategy.GetOverrideCodeMap()
	for typeStr := range override {
		if _, ok := overrideExpectMap[typeStr]; !ok {
			log.Fatalf("expected type for override strategy of type '%s' missing", typeStr)
		}
	}
}

var overrideExpectMap = map[string]string{
	"TIME":    "time.Time",
	"DATE":    "time.Time",
	"UUID":    "string",
	"BOOL":    "bool",
	"VARCHAR": "string",
}

func TestOverrideStrategy_Behavior(t *testing.T) {
	s := strategy.OverrideStrategy{}
	_, err := s.ExecuteStrategy()
	if !errors.As(err, &strategy.UnspecifiedStrategyError{}) {
		t.Fatalf("expected 'UnspecifiedStrategyError', got %v", err)
	}
	err = s.CheckCriteria()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestOverrideStrategy_Implements(t *testing.T) {
	// sample template with an override code

	for types, codeMap := range strategy.GetOverrideCodeMap() {
		for code, creator := range codeMap {
			s := creator()
			if _, ok := s.(strategy.ValueStrategy); ok {
				t.Fatalf("for colummn type '%s' and code '%s': %v", types, code, strategy.ValueStrategyImplementedError{})
			}
		}
	}
}

// ensure that null is treated as override
func TestOverrideStrategy_Null(t *testing.T) {
	genericNullStrategy, err := strategy.GetStrategy("int", "null")
	if err != nil {
		t.Fatal(err)
	}
	if implementsValueStrategy(genericNullStrategy) {
		t.Fatal(strategy.ValueStrategyImplementedError{})
	}
	res, err := genericNullStrategy.ExecuteStrategy()
	if err != nil {
		t.Fatal(err)
	}

	if res != nil {
		t.Fatal(strategy.UnexpectedTypeError{ExpectedType: "nil", ActualType: fmt.Sprintf("%T", res)})
	}
}

func TestOverrideStrategyReturns(t *testing.T) {
	for colType, vals := range strategy.GetOverrideCodeMap() {
		for code, creator := range vals {
			s := creator()
			val, err := s.ExecuteStrategy()
			if err != nil {
				t.Fatal(err)
			}
			if utils.GetStringType(val) != overrideExpectMap[colType] {
				err = strategy.UnexpectedTypeError{ExpectedType: overrideExpectMap[colType], ActualType: utils.GetStringType(val)}
				t.Fatalf("for type '%s' and code '%s': %v", colType, code, err)
			}
		}
	}
}
