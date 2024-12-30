package parameters

type ColumnParser interface {
	ParseColumn() (string, error)
}

var parserMap = map[string]ColumnParser{
	"INT":     &IntColumnParser{},
	"VARCHAR": &VarcharColumnParser{},
	"UUID":    &UUIDParser{},
	"BOOL":    &BooleanParser{},
}

func getColumnParser(dataType string) ColumnParser {
	return parserMap[dataType]
}
