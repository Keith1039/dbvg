package strategy

import (
	"fmt"
	"github.com/Keith1039/dbvg/utils"
	"testing"
)

func implementsValueStrategy(s Strategy) bool {
	_, ok := s.(ValueStrategy)
	return ok
}

var overrideExpectMap = map[string]string{
	"TIME":    "time.Time",
	"DATE":    "time.Time",
	"UUID":    "string",
	"BOOL":    "bool",
	"VARCHAR": "string",
}

func TestOverrideStrategy_Implements(t *testing.T) {
	// sample template with an override code
	for types, codeMap := range overrideCodeMap {
		for code, creator := range codeMap {
			s := creator()
			if _, ok := s.(ValueStrategy); ok {
				t.Fatalf("for colummn type '%s' and code '%s': %v", types, code, ValueStrategyImplementedError{})
			}
		}
	}
}

// ensure that null is treated as override
func TestOverrideStrategy_Null(t *testing.T) {
	genericNullStrategy, err := GetStrategy("int", "null")
	if err != nil {
		t.Fatal(err)
	}
	if implementsValueStrategy(genericNullStrategy) {
		t.Fatal(ValueStrategyImplementedError{})
	}
	res, err := genericNullStrategy.ExecuteStrategy()
	if err != nil {
		t.Fatal(err)
	}

	if res != nil {
		t.Fatal(UnexpectedTypeError{ExpectedType: "nil", ActualType: fmt.Sprintf("%T", res)})
	}
}

func TestOverrideStrategyReturns(t *testing.T) {
	for colType, vals := range overrideCodeMap {
		for code, creator := range vals {
			s := creator()
			val, err := s.ExecuteStrategy()
			if err != nil {
				t.Fatal(err)
			}
			if utils.GetStringType(val) != overrideExpectMap[colType] {
				err = UnexpectedTypeError{ExpectedType: overrideExpectMap[colType], ActualType: utils.GetStringType(val)}
				t.Fatalf("for type '%s' and code '%s': %v", colType, code, err)
			}
		}
	}
}
