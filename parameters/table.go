package parameters

type table struct {
	TableName string
	Columns   []*column
}

type column struct {
	ColumnName string
	Type       string
	Code       int
	Other      string
	Parser     ColumnParser
}
