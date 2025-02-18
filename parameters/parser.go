package parameters

// ColumnParser is an interface that contains the ParseColumn function which is defined by all other parsers
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
		return &UUIDColumnParser{}
	case "BOOL":
		return &BooleanColumnParser{}
	case "DATE":
		return &DateColumnParser{}
	}
	return nil
}
