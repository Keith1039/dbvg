package utils

import (
	"container/list"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	database "github.com/Keith1039/dbvg/db"
	"github.com/dromara/carbon/v2"
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

// MakeTemplates takes in a database connection and an array of tables and formats it into a map suitable for JSON
func MakeTemplates(db *sql.DB, tableOrder []string) map[string]map[string]map[string]any {
	m := make(map[string]map[string]map[string]any)
	relations := database.GetRelationships(db) // get relationships
	for _, tName := range tableOrder {
		m[tName] = makeTemplate(db, tName, relations)
	}
	return m
}

func makeTemplate(db *sql.DB, tName string, relations map[string]map[string]map[string]string) map[string]map[string]any {
	m := make(map[string]map[string]any)
	cols, err := schema.ColumnTypes(db, "", tName)
	if err != nil {
		log.Fatal(err)
	}
	colMap := database.GetColumnMap(db, tName)
	for _, col := range cols {
		_, ok := relations[tName][col.Name()] // check if the column is a fk
		if !ok {
			m[col.Name()] = map[string]any{"type": colMap[col.Name()], "code": "", "value": nil}
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
// along with any errors that indicate problems with the given path. These errors include when a file name is not specified
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
// is a new line for the file. If the file does exist it will be overwritten
func WriteQueriesToFile(path string, queries []string) error {
	// by default this will overwrite existing files
	cleanPath, err := CleanFilePath(path) // make sure the path is clean
	if err != nil {
		return err
	}
	dir, fileName := filepath.Split(cleanPath) // split the dir path and the file name
	if dir != "" {                             // check if the dir path is empty string
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

// WriteInsertTemplateToFile takes in a file path and the InsertTemplate json data. If the file indicated by path
// exist this function will overwrite it with the data in the string array. If the file doesn't exist,
// this function will create it. If the file does exist, it will be overwritten
func WriteInsertTemplateToFile(path string, data map[string]map[string]map[string]any) error {
	// by default this will overwrite existing files
	cleanPath, err := CleanFilePath(path) // make sure the path is clean
	if err != nil {
		return err
	}
	dir, fileName := filepath.Split(cleanPath) // split the dir path and the file name
	if dir != "" {                             // check if the dir path is empty string
		if _, err = os.Stat(dir); os.IsNotExist(err) { // check if directory exists
			err = os.MkdirAll(dir, os.ModePerm) // make all directories and subdirectories
			if err != nil {
				return err // log error and exit
			}
		}
	}
	jsonData, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(dir, fileName), jsonData, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// UpdateInsertTemplate updates an existing JSON template with the most recent information.
// This includes adding new tables, removing irrelevant tables and altering types.
// Shallow verification is done on the given template to ensure that each is valid
// before overwriting data.
//
// Code and value pairs from the old template are still moved into the new template if there
// is a matching entry. This function returns an array of changes made i.e. tables/columns that were added or removed
// alongside any errors.
func UpdateInsertTemplate(path string, newTemplate map[string]map[string]map[string]any) ([]string, error) {
	data, err := RetrieveInsertTemplateJSON(path)
	if err != nil {
		return nil, err
	}
	// we normalize for existing template to see if there's useful data
	for _, columnInfo := range data {
		for _, values := range columnInfo {
			normalizeKeys(values)
		}
	}
	// for the normalization as well
	err = verifyTemplate(newTemplate)
	if err != nil {
		return nil, fmt.Errorf("for the template at '%s' the following error occured: [%w]", path, err)
	}
	changes := updateTemplate(data, newTemplate)
	err = WriteInsertTemplateToFile(path, newTemplate)
	if err != nil {
		return nil, err
	}
	return changes, nil
}

func normalizeKeys(columnInfo map[string]any) {
	var normalizedKey string
	keysToDelete := make(map[string]bool)
	for key, val := range columnInfo {
		normalizedKey = TrimAndLowerString(key)
		if normalizedKey != key {
			columnInfo[normalizedKey] = val
			keysToDelete[key] = true
		}
	}
	for key := range keysToDelete {
		delete(columnInfo, key)
	}
}

func verifyTemplate(m map[string]map[string]map[string]any) error {
	format := "for table '%s' under column '%s' the expected key '%s' is missing"
	// check the keys (doesn't verify the tables or columns yet)
	for tableName, columns := range m {
		for columnName, columnFields := range columns {
			normalizeKeys(columnFields) // trim and lower the keys
			_, ok := columnFields["code"]
			if !ok {
				return fmt.Errorf(format, tableName, columnName, "code")
			}
			_, ok = columnFields["type"]
			if !ok {
				return fmt.Errorf(format, tableName, columnName, "type")
			}
			_, ok = columnFields["value"]
			if !ok {
				return fmt.Errorf(format, tableName, columnName, "value")
			}
		}
	}
	return nil
}

func updateTemplate(oldTemplate map[string]map[string]map[string]any, newTemplate map[string]map[string]map[string]any) []string {
	var val any
	var changesArr []string
	// adding to the new template and getting all the new additions
	for tableName, columns := range newTemplate {
		_, ok := oldTemplate[tableName]
		if !ok {
			changesArr = append(changesArr, fmt.Sprintf("+ new table '%s' added", tableName))
		} else {
			for columnName := range columns {
				_, ok = oldTemplate[tableName][columnName]
				if ok {
					// assume that new template is given via MakeTemplates() and thus should have the correct type

					if val, ok = oldTemplate[tableName][columnName]["code"]; ok {
						newTemplate[tableName][columnName]["code"] = val
					}
					if val, ok = oldTemplate[tableName][columnName]["value"]; ok {
						newTemplate[tableName][columnName]["value"] = val
					}
				} else {
					changesArr = append(changesArr, fmt.Sprintf("+ new column '%s' added to table '%s'", columnName, tableName))
				}
			}
		}
	}
	// now to get the deletions
	for tableName, columns := range oldTemplate {
		if _, ok := newTemplate[tableName]; !ok {
			changesArr = append(changesArr, fmt.Sprintf("- table '%s' removed", tableName))
		} else {
			for columnName := range columns {
				if _, ok = newTemplate[tableName][columnName]; !ok {
					changesArr = append(changesArr, fmt.Sprintf("- column '%s' removed from table '%s'", columnName, tableName))
				}
			}
		}
	}
	return changesArr
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
	return c.StdTime(), nil
}

// RetrieveInsertTemplateJSON takes in a path to a file and retrieves the JSON data for an
// InsertTemplate. Various checks occur to ensure that the path given works on any operating system
// so long as it is a valid path.
func RetrieveInsertTemplateJSON(path string) (map[string]map[string]map[string]any, error) {
	var bytes []byte
	data := make(map[string]map[string]map[string]any) // data container
	cleanFilePath, err := CleanFilePath(path)          // clean the file path and check for errors
	if err != nil {                                    // error check
		return nil, err
	}
	if _, err = os.Stat(cleanFilePath); os.IsNotExist(err) { // check if file exists
		return nil, err
	}
	bytes, err = os.ReadFile(cleanFilePath) // read the bytes from the file
	if err != nil {                         // error check
		return nil, err
	}
	err = json.Unmarshal(bytes, &data) // unmarshall bytes into container
	if err != nil {                    // error check
		return nil, err
	}
	return data, nil // return the data for validation
}
