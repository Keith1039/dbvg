package graph

import (
	"container/list"
	"fmt"
)

// MissingTableError is a custom error that is created when a given table can not be found in the database
type MissingTableError struct {
	tableName string
}

// Error returns a formated string that informs the user that the table does not exist in the database
func (e MissingTableError) Error() string {
	return fmt.Sprintf("Table '%s' does not exist in database", e.tableName)
}

// CyclicError is a custom error that is created when cycles are detected in the database schema
type CyclicError struct {
	cycles *list.List
}

// Error returns a formatted string informing the user of all detected cycles in the database
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
