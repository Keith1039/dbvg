package strategy_test

import (
	"errors"
	"github.com/Keith1039/dbvg/strategy"
	"testing"
)

func TestUnspecifiedParameters(t *testing.T) {
	// check with optional strategy
	testOpt := &strategy.OptionalStrategy{}
	err := testOpt.CheckCriteria()
	if !errors.As(err, &strategy.UnspecifiedCriteriaError{}) {
		t.Fatalf("expected UnspecifiedCriteriaError received '%v'", err)
	}

	_, err = testOpt.ExecuteStrategy()
	if !errors.As(err, &strategy.UnspecifiedStrategyError{}) {
		t.Fatalf("expected UnspecifiedStrategyError received '%v'", err)
	}

	// check with required strategy
	testReq := &strategy.RequiredStrategy{}
	testReq.Value = 5 // need to set it to non-nil value
	err = testReq.CheckCriteria()
	if !errors.As(err, &strategy.UnspecifiedCriteriaError{}) {
		t.Fatalf("expected UnspecifiedCriteriaError received '%v'", err)
	}

	_, err = testReq.ExecuteStrategy()
	if !errors.As(err, &strategy.UnspecifiedStrategyError{}) {
		t.Fatalf("expected UnspecifiedStrategyError received '%v'", err)
	}
}
