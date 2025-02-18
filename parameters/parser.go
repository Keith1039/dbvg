package parameters

type ColumnParser interface {
	ParseColumn(col column) (string, error)
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
