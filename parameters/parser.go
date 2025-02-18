package parameters

type ColumnParser interface {
	ParseColumn(col column) (string, error)
}

var parserMap = map[string]ColumnParser{
	"INT":     &IntColumnParser{},
	"VARCHAR": &VarcharColumnParser{},
	"UUID":    &UUIDParser{},
	"BOOL":    &BooleanParser{},
}

func getColumnParser(dataType string) ColumnParser {
	switch dataType {
	case "INT":
		return &IntColumnParser{}
	case "VARCHAR":
		return &VarcharColumnParser{}
	case "UUID":
		return &UUIDParser{}
	case "BOOL":
		return &BooleanParser{}
	}
	return nil
}
