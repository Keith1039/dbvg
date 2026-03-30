package parameters

import (
	"database/sql"
	"github.com/Keith1039/dbvg/template"
)

type table struct {
	TableName string
	Columns   []*column
}

type column struct {
	ColumnName    string
	ColumnDetails *sql.ColumnType
	Type          string
	Pair          template.StrategyCodePair
}
