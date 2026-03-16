package strategy

import (
	"fmt"
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/utils"
)

type factories interface {
	func() Strategy | func() ValueStrategy
}

var overrideCodeMap = map[string]map[string]func() Strategy{
	"DATE": {"NOW": NewNowStrategy},
	"UUID": {"UUID": NewUUIDStrategy},
	"BOOL": {"RANDOM": NewRandomBoolStrategy},
	"VARCHAR": {
		"EMAIL":     NewEmailVarcharStrategy,
		"FIRSTNAME": NewFirstNameVarcharStrategy,
		"LASTNAME":  NewLastNameVarcharStrategy,
		"FULLNAME":  NewFullNameVarcharStrategy,
		"PHONE":     NewPhoneVarcharStrategy,
		"COUNTRY":   NewCountryVarcharStrategy,
		"ADDRESS":   NewAddressVarcharStrategy,
		"ZIPCODE":   NewZipCodeVarcharStrategy,
		"CITY":      NewCityVarcharStrategy,
	},
}

var optionalCodeMap = map[string]map[string]func() ValueStrategy{
	"INT": {"SERIAL": NewSerialStrategy},
}

var requiredCodeMap = map[string]map[string]func() ValueStrategy{
	"BOOL": {
		"STATIC": NewStaticBoolStrategy,
	},
	"DATE": {
		"RANDOM": NewRandomDateStrategy,
		"STATIC": NewStaticDateStrategy,
	},
	"INT": {
		"RANDOM": NewRandomIntStrategy,
		"STATIC": NewStaticIntStrategy,
	},
	"FLOAT": {
		"RANDOM": NewRandomFloatStrategy,
		"STATIC": NewStaticFloatStrategy,
	},
	"VARCHAR": {
		"STATIC": NewStaticVarcharStrategy,
		"REGEX":  NewRegexStrategy,
	},
}

// GetStrategy returns a strategy from the internal maps based on the columnType and codeName given.
// It also returns any errors that occurred in the attempt
func GetStrategy(columnType, codeName string) (Strategy, error) {
	columnType = utils.TrimAndUpperString(columnType)
	if !database.IsSupportedType(columnType) {
		// TODO: make this an actual error
		return nil, fmt.Errorf("unsupported column type: %s", columnType)
	}
	codeName = utils.TrimAndUpperString(codeName)
	if codeName == "NULL" {
		switch columnType {
		case "UUID":
			return NewNullUUIDStrategy(), nil
		default:
			return NewNullStrategy(), nil
		}
	}

	strategy, ok := overrideCodeMap[columnType][codeName]
	if ok {
		return strategy(), nil
	}
	valStrategy, ok := optionalCodeMap[columnType][codeName]
	if ok {
		return valStrategy(), nil
	}
	valStrategy, ok = requiredCodeMap[columnType][codeName]
	if ok {
		return valStrategy(), nil
	}
	return nil, UnsupportedCodeError{ColumnType: columnType, Code: codeName}
}

func genericCheckForStrategyThenAdd[T factories](strategyMap map[string]map[string]T, columnType string, codeName string, strategy T) error {
	columnType = utils.TrimAndUpperString(columnType)
	codeName = utils.TrimAndUpperString(codeName)
	if strategy == nil {
		return RequiredParameterNilError{Name: "strategy"}
	}
	if !database.IsSupportedType(columnType) {
		// TODO: make this an actual error
		return fmt.Errorf("unsupported column type: %s", columnType)
	}
	if codeName == "NULL" { // make sure people can't try to overwrite NULL code
		return ExistingStrategyError{ColumnType: columnType, Code: codeName}
	}
	s, _ := GetStrategy(columnType, codeName) // check if an existing strategy is defined
	if s != nil {
		return ExistingStrategyError{ColumnType: columnType, Code: codeName}
	}
	if _, ok := strategyMap[columnType]; ok {
		strategyMap[columnType][codeName] = strategy
	} else {
		strategyMap[columnType] = map[string]T{codeName: strategy}
	}
	return nil

}

// AddNewOverrideStrategy adds a strategy function to the internal map, an error is returned if the process fails
func AddNewOverrideStrategy(columnType string, codeName string, strategy func() Strategy) error {
	if strategy == nil {
		return RequiredParameterNilError{Name: "strategy"}
	}
	s := strategy()
	if _, ok := s.(ValueStrategy); ok {
		return ValueStrategyImplementedError{}
	}
	return genericCheckForStrategyThenAdd[func() Strategy](overrideCodeMap, columnType, codeName, strategy)
}

// AddNewOptionalStrategy adds a strategy function to the internal map, an error is returned if the process fails
func AddNewOptionalStrategy(columnType string, codeName string, strategy func() ValueStrategy) error {
	return genericCheckForStrategyThenAdd[func() ValueStrategy](optionalCodeMap, columnType, codeName, strategy)
}

// AddNewRequiredStrategy adds a strategy function to the internal map, an error is returned if the process fails
func AddNewRequiredStrategy(columnType string, codeName string, strategy func() ValueStrategy) error {
	return genericCheckForStrategyThenAdd[func() ValueStrategy](requiredCodeMap, columnType, codeName, strategy)
}
