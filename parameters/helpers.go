package parameters

import (
	"container/list"
	"fmt"
	"github.com/Keith1039/dbvg/utils"
)

func getAllColumnNames(columns []*column) []string {
	l := list.New()
	for _, col := range columns {
		l.PushBack(col.ColumnName)
	}
	return utils.ListToStringArray(l)
}

func createParameterString(length int) []string {
	arr := make([]string, length)
	for i := 0; i < length; i++ {
		arr[i] = fmt.Sprintf("$%d", i+1)
	}
	return arr
}

func createDeleteQuery(allColumn, parameterString []string) []string {
	l := list.New()
	for i, col := range allColumn {
		l.PushBack(fmt.Sprintf("%s=%s", col, parameterString[i]))
	}
	return utils.ListToStringArray(l)
}
