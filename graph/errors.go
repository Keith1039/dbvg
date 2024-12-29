package graph

import "fmt"

type CyclicError struct {
	tableName  string
	rTableName string
}

func (e CyclicError) Error() string {
	return fmt.Sprintf("Circular dependency between tables %s and %s detected", e.tableName, e.rTableName)
}
