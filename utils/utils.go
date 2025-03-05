package utils

import (
	"container/list"
	"database/sql"
	database "github.com/Keith1039/dbvg/db"
	"github.com/jimsmart/schema"
	"log"
)

// ListToStringArray takes in a linked list and returns it as a string array
func ListToStringArray(l *list.List) []string {
	// converts a linkedlist into a string array
	arr := make([]string, l.Len())
	node := l.Front()
	slider := 0
	for node != nil {
		arr[slider] = node.Value.(string) // set array val
		slider++                          // increment slider
		node = node.Next()                // move to the next node
	}
	return arr // return the array
}

// MakeTemplates takes in a database connection and an array of tables and formats it into a map suitable for JSON
func MakeTemplates(db *sql.DB, tableOrder []string) map[string]map[string]map[string]string {
	m := make(map[string]map[string]map[string]string)
	relations := database.GetRelationships(db) // get relationships
	for _, tName := range tableOrder {
		m[tName] = makeTemplate(db, tName, relations)
	}
	return m
}

func makeTemplate(db *sql.DB, tName string, relations map[string]map[string]map[string]string) map[string]map[string]string {
	m := make(map[string]map[string]string)
	cols, err := schema.ColumnTypes(db, "", tName)
	colMap := database.GetColumnMap(db, tName)
	if err != nil {
		log.Fatal(err)
	}
	for _, col := range cols {
		_, ok := relations[tName][col.Name()] // check if the column is a fk
		if !ok {
			m[col.Name()] = map[string]string{"Type": colMap[col.Name()], "Code": "", "Value": ""}
		}
	}
	return m
}
