package utils

import (
	"container/list"
	"database/sql"
	"errors"
	"fmt"
	database "github.com/Keith1039/dbvg/db"
	"github.com/golang-module/carbon"
	"github.com/jimsmart/schema"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// generic function for turning a linked list to an array
func listToArray[T any](l *list.List) []T {
	// converts a linkedlist into an array
	arr := make([]T, l.Len())
	node := l.Front()
	slider := 0
	for node != nil {
		arr[slider] = node.Value.(T) // set array val
		slider++                     // increment slider
		node = node.Next()           // move to the next node
	}
	return arr // return the array
}

// ListToStringArray takes in a linked list and returns it as a string array assuming each element in the linked list is a string
func ListToStringArray(l *list.List) []string {
	// converts a linkedlist into a string array
	return listToArray[string](l)
}

// ListToAnyArray takes in a linked list and returns it as an any array
func ListToAnyArray(l *list.List) []any {
	return listToArray[any](l)
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

// TrimAndUpperString trims space and makes each character upper case for the given string
func TrimAndUpperString(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}

// CleanFilePath ensures that the file path is a proper file path before returning an OS specific path using `filepath.clean()`
// along with any errors that indicate problems with the given path
func CleanFilePath(path string) (string, error) {
	testPath := strings.TrimSpace(path)
	if testPath == "" {
		return "", errors.New("path is empty")
	} else if filepath.Dir(testPath) == filepath.Clean(testPath) { // the two strings shouldn't be equal if there's a filepath specified
		return "", errors.New("no file name specified")
	}
	path = filepath.Clean(testPath) // clean the path
	return path, nil
}

// WriteQueriesToFile takes in a file path and an array of strings. If the file indicated by path
// exist this function will overwrite it with the data in the string array. If the file doesn't exist,
// this function will create it before inputting the string array data. Each index of the string array
// is a new line for the file.
func WriteQueriesToFile(path string, queries []string) error {
	// by default this will overwrite existing files
	cleanPath, err := CleanFilePath(path) // make sure the path is clean
	if err != nil {
		return err
	}
	dir, fileName := filepath.Split(cleanPath) // split the dir path and the file name
	if fileName == "" {                        // check to see if there is a valid file name
		return errors.New("file name not specified") // error out
	}
	if dir != "" { // check if the dir path is empty string
		if _, err = os.Stat(dir); os.IsNotExist(err) { // check if directory exists
			err = os.MkdirAll(dir, os.ModePerm) // make all directories and subdirectories
			if err != nil {
				return err // log error and exit
			}
		}
	}
	file, err := os.Create(filepath.Join(dir, fileName)) // create the file name
	if err != nil {                                      // error check
		return err
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file) // close the file
	return writeToFile(file, queries) // write the queries to the file
}

func writeToFile(file *os.File, queries []string) error {
	// edge cases to avoid appending any sort of junk data to file
	if len(queries) == 0 || (len(queries) == 1 && queries[0] == "") {
		return nil
	}
	// writes all the queries to the file
	for i, query := range queries {
		var err error
		if i == len(queries)-1 { // last index
			_, err = file.WriteString(query) // write to file without line separator
		} else {
			_, err = file.WriteString(fmt.Sprintf("%s\n", query)) // write to file
		}
		if err != nil { // error check
			return err
		}
	}
	err := file.Sync() // synch file
	if err != nil {    // error check
		return err
	}
	return nil
}

// GetStringType is a function take takes in a value of any type and returns the string name of the type of value given
func GetStringType(val any) string {
	return fmt.Sprintf("%T", val)
}

// GetTimeFromString takes in a string and attempts to parse it using carbon.
// if no errors occurs, it returns the time.Time version of the parsed string
// otherwise it returns an empty time.Time struct along with the error that occurred
func GetTimeFromString(dateString string) (time.Time, error) {
	dateString = strings.TrimSpace(dateString)
	c := carbon.Parse(dateString)
	if c.Error != nil {
		return time.Time{}, c.Error
	}
	return c.ToStdTime(), nil
}
