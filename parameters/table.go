package parameters

import "database/sql"

type table struct {
	TableName string
	Columns   []*column
}

type column struct {
	ColumnName    string
	ColumnDetails *sql.ColumnType
	Type          string
	Code          int
	Other         string
	Parser        ColumnParser
}
