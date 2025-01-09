package graph

import (
	"container/list"
	"fmt"
)

type MissingTableError struct {
	tableName string
}

func (e MissingTableError) Error() string {
	return fmt.Sprintf("Table '%s' does not exist in database", e.tableName)
}

type CyclicError struct {
	cycles *list.List
}

func (e CyclicError) Error() string {
	cycleString := "error, the following cycles have been detected in the database schema: "
	node := e.cycles.Front()
	for node != nil {
		cycleString += node.Value.(string)
		if node.Next() != nil {
			cycleString += " | "
		}
		node = node.Next()
	}
	return cycleString
}
