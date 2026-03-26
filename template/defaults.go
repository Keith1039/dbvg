package template

import "github.com/Keith1039/dbvg/strategy"

var defaults = map[string]func() strategy.Strategy{
	"INT":     func() strategy.Strategy { return strategy.NewSerialStrategy() },
	"FLOAT":   defaultFloatRandom,
	"UUID":    func() strategy.Strategy { return strategy.NewUUIDStrategy() },
	"DATE":    func() strategy.Strategy { return strategy.NewNowStrategy() },
	"BOOL":    func() strategy.Strategy { return strategy.NewRandomBoolStrategy() },
	"VARCHAR": defaultRegex,
}

func defaultFloatRandom() strategy.Strategy {
	s := strategy.NewRandomFloatStrategy()
	s.SetValue([]int{1, 10})
	return s
}

func defaultRegex() strategy.Strategy {
	s := strategy.NewRegexStrategy()
	s.SetValue("[a-zA-Z]{10}")
	return s
}
