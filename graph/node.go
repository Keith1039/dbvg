package graph

type TableInfoNode struct {
	tableName       string
	parentTableName string
}

type TableNode struct {
	TableName  string
	ColumnData map[string]string
	Parameters map[string]string
}
