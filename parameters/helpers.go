package parameters

import (
	"fmt"
)

func getQueryParams(columns []*column) ([]string, []string, []string) {
	colLen := len(columns)
	allColumns := make([]string, colLen)
	paramStrings := make([]string, colLen)
	deleteQuery := make([]string, colLen)
	for i, col := range columns {
		allColumns[i] = col.ColumnName
		paramStrings[i] = fmt.Sprintf("$%d", i+1)
		deleteQuery[i] = fmt.Sprintf("%s=%s", allColumns[i], paramStrings[i])
	}
	return allColumns, paramStrings, deleteQuery
}
