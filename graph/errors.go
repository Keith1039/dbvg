package graph

import "fmt"

type MissingTableError struct {
	tableName string
}

func (e MissingTableError) Error() string {
	return fmt.Sprintf("Table %s does not exist in database", e.tableName)
}

type CyclicError struct {
	tableName  string
	rTableName string
}

func (e CyclicError) Error() string {
	return fmt.Sprintf("Circular dependency between tables %s and %s detected", e.tableName, e.rTableName)
}
