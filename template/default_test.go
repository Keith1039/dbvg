package template_test

import (
	"github.com/Keith1039/dbvg/strategy"
	"github.com/Keith1039/dbvg/template"
	"reflect"
	"testing"
)

func TestGetDefault(t *testing.T) {
	defaults := template.GetDefaults()
	defaults["NEW TYPE"] = func() strategy.Strategy { return &strategy.OverrideStrategy{} }
	defaults2 := template.GetDefaults()
	if reflect.DeepEqual(defaults, defaults2) {
		t.Fatal("changes made to the clone should not affect the internal values")
	}
}

// check if executing default strategies can be executed as is
func TestGetDefaultExec(t *testing.T) {
	defaults := template.GetDefaults()
	for typeString, stratFunc := range defaults {
		strat := stratFunc()
		_, err := strat.ExecuteStrategy()
		if err != nil {
			t.Fatalf("default strategy of type '%s' failed with error '%v'", typeString, err)
		}
	}
}

func TestGetDefaultCodes(t *testing.T) {
	defaults := template.GetDefaultCodes()
	defaults["INT"] = "NULL"
	defaults2 := template.GetDefaultCodes()
	if reflect.DeepEqual(defaults, defaults2) {
		t.Fatal("changes made to the clone should not affect the internal values")
	}
}

// a strategy is the same if it has the same strategy and criteria function
//func strategyIsEqual(stratA strategy.Strategy, stratB strategy.Strategy) bool {
//	override1, ok := stratA.(*strategy.OverrideStrategy)
//	override2, ok2 := stratB.(*strategy.OverrideStrategy)
//	if ok && ok2 {
//		// check if they point to te same strategy
//		return reflect.DeepEqual(override1, override2)
//	}
//
//	optional1, ok := stratA.(*strategy.OptionalStrategy)
//	optional2, ok2 := stratB.(*strategy.OptionalStrategy)
//	if ok && ok2 {
//		return reflect.DeepEqual(optional1, optional2)
//	}
//
//	required1, ok := stratA.(*strategy.RequiredStrategy)
//	required2, ok2 := stratB.(*strategy.RequiredStrategy)
//	if ok && ok2 {
//		return reflect.DeepEqual(required1, required2)
//	}
//	return false
//}

//func TestCodeAndStrategyAlignment(t *testing.T) {
//	defaultStrategies := template.GetDefaults()
//	defaultCodes := template.GetDefaultCodes()
//	for typeStr, code := range defaultCodes {
//		stratFunc, ok := defaultStrategies[typeStr]
//		if !ok {
//			t.Fatalf("missing strategy for type '%s'", typeStr)
//		}
//		strat := stratFunc()
//		expectedStrat, err := strategy.GetStrategy(typeStr, code)
//		if err != nil {
//			t.Fatal(err)
//		}
//		if !strategyIsEqual(strat, expectedStrat) {
//			t.Fatalf("strategy of type '%s' for code '%s' doesn't match expected", typeStr, code)
//		}
//
//	}
//}
