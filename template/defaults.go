package template

import (
	"github.com/Keith1039/dbvg/strategy"
	"maps"
)

var DEFAULTREGEX = "[a-zA-Z]{10}"

var defaults = map[string]func() strategy.Strategy{
	"INT":     func() strategy.Strategy { return defaultSerial() },
	"FLOAT":   defaultFloatRandom,
	"UUID":    func() strategy.Strategy { return strategy.NewUUIDStrategy() },
	"DATE":    func() strategy.Strategy { return strategy.NewNowDateStrategy() },
	"TIME":    func() strategy.Strategy { return strategy.NewNowTimeStrategy() },
	"BOOL":    func() strategy.Strategy { return strategy.NewRandomBoolStrategy() },
	"VARCHAR": defaultRegex,
}

// GetDefaults returns a clone of the internal defaults map linking the types to their
// Strategies. These Strategies can be executed as is
func GetDefaults() map[string]func() strategy.Strategy {
	return maps.Clone(defaults)
}

var defaultCode = map[string]string{
	"INT":     "SERIAL",
	"FLOAT":   "RANDOM",
	"UUID":    "UUID",
	"DATE":    "NOW",
	"TIME":    "NOW",
	"BOOL":    "RANDOM",
	"VARCHAR": "REGEX",
}

// GetDefaultCodes returns a clone of the internal defaultCode map linking a type
// to the code represented in the defaults map.
func GetDefaultCodes() map[string]string {
	return maps.Clone(defaultCode)
}

func defaultFloatRandom() strategy.Strategy {
	s := strategy.NewRandomFloatStrategy()
	s.SetValue([]float64{1, 10})
	return s
}

func defaultRegex() strategy.Strategy {
	s := strategy.NewRegexStrategy()
	s.SetValue(DEFAULTREGEX)
	return s
}

func defaultSerial() strategy.Strategy {
	s := strategy.NewSerialStrategy()
	s.SetValue(1)
	return s
}
