package utils

import (
	"container/list"
	"database/sql"
	"fmt"
	database "github.com/Keith1039/dbvg/db"
	"github.com/jimsmart/schema"
	"log"
	"os"
	"path/filepath"
	"strings"
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

// TrimAndLowerString trims space and lowers the given string
func TrimAndLowerString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func WriteQueriesToFile(path string, queries []string) {
	dir, fileName := filepath.Split(path) // split the dir path and the file name
	if dir != "" {                        // check if the dir path is empty string
		if _, err := os.Stat(dir); os.IsNotExist(err) { // check if directory exists
			err = os.MkdirAll(dir, os.ModePerm) // make all directories and subdirectories
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	file, err := os.Create(fileName) // create the file name
	defer file.Close()               // close the file
	if err != nil {                  // error check
		log.Fatal(err)
	}
	writeToFile(file, queries) // write the queries to the file
}

func writeToFile(file *os.File, queries []string) {
	// writes all the queries to the file
	for _, query := range queries {
		_, err := fmt.Fprintln(file, query)
		if err != nil {
			log.Fatal(err)
		}
	}
}
