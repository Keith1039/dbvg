package strategy

import (
	"fmt"
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/utils"
	"maps"
)

type factories interface {
	func() Strategy | func() ValueStrategy
}

var defaults = map[string]map[string]bool{
	"DATE": {
		"NOW":    true,
		"RANDOM": true,
		"STATIC": true,
	},
	"UUID": {"UUID": true},
	"BOOL": {
		"RANDOM": true,
		"STATIC": true,
	},
	"VARCHAR": {
		"EMAIL":     true,
		"FIRSTNAME": true,
		"LASTNAME":  true,
		"FULLNAME":  true,
		"PHONE":     true,
		"COUNTRY":   true,
		"ADDRESS":   true,
		"ZIPCODE":   true,
		"CITY":      true,
		"STATIC":    true,
		"REGEX":     true,
	},
	"INT": {
		"SERIAL": true,
		"RANDOM": true,
		"STATIC": true,
	},
	"FLOAT": {
		"RANDOM": true,
		"STATIC": true,
	},
}

func makeCopyAndReturn[T factories](m map[string]map[string]T) map[string]map[string]T {
	dupe := maps.Clone(m)
	return dupe
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

// GetOverrideCodeMap returns a copy of the internal map for codes of the OverrideStrategy type
func GetOverrideCodeMap() map[string]map[string]func() Strategy {
	return makeCopyAndReturn(overrideCodeMap)
}

var optionalCodeMap = map[string]map[string]func() ValueStrategy{
	"INT": {"SERIAL": NewSerialStrategy},
}

// GetOptionalCodeMap returns a copy of the internal map for codes of the OptionalStrategy type
func GetOptionalCodeMap() map[string]map[string]func() ValueStrategy {
	return makeCopyAndReturn(optionalCodeMap)
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

// GetRequiredCodeMap returns a copy of the internal map for codes of the RequiredStrategy type
func GetRequiredCodeMap() map[string]map[string]func() ValueStrategy {
	return makeCopyAndReturn(requiredCodeMap)
}

func deleteFromMap[T factories](m map[string]map[string]T, columnType, codeName string) {
	// we delete the code from the nested map
	delete(m[columnType], codeName)
	// we check if that code was the last
	if len(m[columnType]) == 0 {
		delete(m, columnType) // if it was, clean up the empty map
	}
}

// DeleteStrategy deletes a given strategy from the internal maps and returns any errors.
// This function cannot affect the default codes, i.e. the codes hardcoded into the library and will return an error if this is attempted.
// If the code columnType and codename already don't exist in a map, this is considered a no-op
func DeleteStrategy(columnType, codeName string) error {
	columnType = utils.TrimAndUpperString(columnType)
	codeName = utils.TrimAndUpperString(codeName)
	if codeName == "NULL" { // null is predefined for everyone
		return nil
	}
	s, _ := GetStrategy(columnType, codeName) // we don't care if there's an error, we care that we got a strategy
	if s != nil {
		if _, ok := defaults[columnType][codeName]; ok {
			return DeleteDefaultError{columnType: columnType, code: codeName}
		} else {
			// literally has to be one of these, if not it would return nil
			_, ok = overrideCodeMap[columnType][codeName]
			if ok {
				deleteFromMap(overrideCodeMap, columnType, codeName)
				return nil
			}
			_, ok = optionalCodeMap[columnType][codeName]
			if ok {
				deleteFromMap(optionalCodeMap, columnType, codeName)
				return nil
			}
			_, ok = requiredCodeMap[columnType][codeName]
			if ok {
				deleteFromMap(requiredCodeMap, columnType, codeName)
				return nil
			}
		}
	}
	return nil
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
