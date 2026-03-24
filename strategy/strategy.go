package strategy

type Strategy interface {
	ExecuteStrategy() (any, error)
	CheckCriteria() error
}

type ValueStrategy interface {
	Strategy
	SetValue(any)
}

// defaultStrategy is an embedded struct meant to be used by the main types
type defaultStrategy struct {
	Value    any
	Strategy func(val any) (any, error)
	Criteria func(val any) error
}

func (s *defaultStrategy) ExecuteStrategy() (any, error) {
	if s.Strategy == nil {
		return nil, UnspecifiedStrategyError{}
	}
	return s.Strategy(s.Value)
}

func (s *defaultStrategy) CheckCriteria() error {
	if s.Criteria == nil {
		return UnspecifiedCriteriaError{}
	}
	return s.Criteria(s.Value)
}

func (s *defaultStrategy) SetValue(val any) {
	s.Value = val
}

type CustomStrategy struct {
}

func (s *CustomStrategy) ExecuteStrategy() (any, error) {
	return nil, nil
}

func (s *CustomStrategy) CheckCriteria() error {
	return nil
}

func (s *CustomStrategy) SetValue(val any) {

}
