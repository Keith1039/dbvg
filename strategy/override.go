package strategy

import (
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"golang.org/x/exp/rand"
	"time"
)

// OverrideStrategy is a type of Strategy that doesn't take in a value and simply executes it's given strategy function
type OverrideStrategy struct {
	Strategy func() (any, error)
}

// ExecuteStrategy for the OverrideStrategy executes and returns it's strategy.
// if the strategy is undefined, i.e. nil, the program logs the failure and exits using log.Fatal
func (s *OverrideStrategy) ExecuteStrategy() (any, error) {
	// override strategies don't care about values
	// so there's no need to check criteria
	if s.Strategy == nil {
		return nil, UnspecifiedStrategyError{}
	}
	return s.Strategy()
}

// CheckCriteria for the OverrideStrategy returns nil as OverrideStrategy's do not take in a value to evaluate
func (s *OverrideStrategy) CheckCriteria() error {
	// Override strategies don't care about their values
	// so there's nothing to check
	return nil
}

// overrides don't have criteria defined

// handles null code
func handleNullCode() (any, error) {
	return nil, nil
}

// NewNullStrategy defines and returns a Strategy of type OverrideStrategy meant to handle the "NUll" code for all types except "UUID"
func NewNullStrategy() Strategy {
	return &OverrideStrategy{Strategy: handleNullCode}
}

func handleNullCodeUUID() (any, error) {
	return uuid.Nil.String(), nil
}

// NewNullUUIDStrategy defines and returns a Strategy of type OverrideStrategy meant to handle the "NUll" code for type "UUID"
func NewNullUUIDStrategy() Strategy {
	return &OverrideStrategy{Strategy: handleNullCodeUUID}
}

func handleNowCode() (any, error) {
	return time.Now().String()[0:19], nil
}

// NewNowStrategy defines and returns a Strategy of type OverrideStrategy meant to handle code "NOW" for type "DATE"
func NewNowStrategy() Strategy {
	return &OverrideStrategy{Strategy: handleNowCode}
}

func handleUUIDCode() (any, error) {
	return uuid.New().String(), nil
}

// NewUUIDStrategy defines and returns a Strategy of type OverrideStrategy meant to handle code "UUID" for type "UUID"
func NewUUIDStrategy() Strategy {
	return &OverrideStrategy{Strategy: handleUUIDCode}
}

func handleRandomCodeBool() (any, error) {
	num := rand.Intn(100)
	var res bool
	if num < 50 {
		res = true
	} else {
		res = false
	}
	return res, nil
}

// NewRandomBoolStrategy defines and returns a Strategy of type OverrideStrategy meant to handle code "RANDOM" for type "BOOL"
func NewRandomBoolStrategy() Strategy {
	return &OverrideStrategy{Strategy: handleRandomCodeBool}
}

func handleEmailCode() (any, error) {
	return gofakeit.Email(), nil
}

// NewEmailVarcharStrategy defines and returns a Strategy of type OverrideStrategy meant to handle code "EMAIL" for type "VARCHAR"
func NewEmailVarcharStrategy() Strategy {
	return &OverrideStrategy{Strategy: handleEmailCode}
}

func handleFirstName() (any, error) {
	return gofakeit.Person().FirstName, nil
}

// NewFirstNameVarcharStrategy defines and returns a Strategy of type OverrideStrategy meant to handle code "FIRSTNAME" for type "VARCHAR"
func NewFirstNameVarcharStrategy() Strategy {
	return &OverrideStrategy{Strategy: handleFirstName}
}

func handleLastName() (any, error) {
	return gofakeit.Person().LastName, nil
}

// NewLastNameVarcharStrategy defines and returns a Strategy of type OverrideStrategy meant to handle code "LASTNAME" for type "VARCHAR"
func NewLastNameVarcharStrategy() Strategy {
	return &OverrideStrategy{Strategy: handleLastName}
}

func handleFullName() (any, error) {
	return gofakeit.Name(), nil
}

// NewFullNameVarcharStrategy defines and returns a Strategy of type OverrideStrategy meant to handle code "FULLNAME" for type "VARCHAR"
func NewFullNameVarcharStrategy() Strategy {
	return &OverrideStrategy{Strategy: handleFullName}
}

func handlePhone() (any, error) {
	return gofakeit.Phone(), nil
}

// NewPhoneVarcharStrategy defines and returns a Strategy of type OverrideStrategy meant to handle code "PHONE" for type "VARCHAR"
func NewPhoneVarcharStrategy() Strategy {
	return &OverrideStrategy{Strategy: handlePhone}
}

func handleCountry() (any, error) {
	return gofakeit.Country(), nil
}

// NewCountryVarcharStrategy defines and returns a Strategy of type OverrideStrategy meant to handle code "COUNTRY" for type "VARCHAR"
func NewCountryVarcharStrategy() Strategy {
	return &OverrideStrategy{Strategy: handleCountry}
}

func handleAddress() (any, error) {
	return gofakeit.Address().Address, nil
}

// NewAddressVarcharStrategy defines and returns a Strategy of type OverrideStrategy meant to handle code "ADDRESS" for type "VARCHAR"
func NewAddressVarcharStrategy() Strategy {
	return &OverrideStrategy{Strategy: handleAddress}
}

func handleZipCode() (any, error) {
	return gofakeit.Zip(), nil
}

// NewZipCodeVarcharStrategy defines and returns a Strategy of type OverrideStrategy meant to handle code "ZIPCODE" for type "VARCHAR"
func NewZipCodeVarcharStrategy() Strategy {
	return &OverrideStrategy{Strategy: handleZipCode}
}

func handleCity() (any, error) {
	return gofakeit.City(), nil
}

// NewCityVarcharStrategy defines and returns a Strategy of type OverrideStrategy meant to handle code "CITY" for type "VARCHAR"
func NewCityVarcharStrategy() Strategy {
	return &OverrideStrategy{Strategy: handleCity}
}
