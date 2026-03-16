package template

import (
	"github.com/Keith1039/dbvg/graph"
	"github.com/Keith1039/dbvg/strategy"
	"github.com/Keith1039/dbvg/utils"
)

var insertSchema = map[string]string{
	"code":  "string",
	"type":  "string",
	"value": "any",
}

// InsertTemplate is a struct that handles interactions with an Insert JSON template
type InsertTemplate struct {
	data map[string]map[string]map[string]any
}

func (t *InsertTemplate) TemplateFrom(tableData map[string]map[string]string, path string) error {
	data, err := retrieveJSON(path) // run some validation and retrieve JSON data
	if err != nil {                 // error check
		return err
	}
	err = t.validateTemplate(data, tableData, insertSchema) // validate the data given to see if it matches schema
	// add an extra step for checking if the value has a type that is supported by the code
	if err != nil { // error check
		return err
	}
	t.data = data // set the data if there's no error
	return nil
}

// validateTemplate confirms that the information gained from a JSON file is valid through a series of checks
func (t *InsertTemplate) validateTemplate(jsonData map[string]map[string]map[string]any, tableData map[string]map[string]string, schema map[string]string) error {
	var typeMap map[string]string
	for tableName, columns := range jsonData { // loop through each key
		if _, ok := tableData[tableName]; !ok { // check if the tableName exists in the schema
			return graph.NewMissingTableError(tableName)
		}
		for colName, columnInfo := range columns {
			if _, ok := tableData[tableName][colName]; !ok { // check if the column exists in that schema
				return graph.NewMissingColumnError(tableName, colName)
			}
			normalizeKeys(columnInfo)                  // trims and lowers each key while maintaining the key value pairs
			typeMap = makeTypeMap(columnInfo)          // check the types for the key value pairs in the template
			err := checkAgainstSchema(typeMap, schema) // check the type map against the schema
			if err != nil {
				return wrapError(tableName, colName, err)
			}
			normalizeType(columnInfo) // ensures the convention of the type field
			err = checkExpectedType(tableData[tableName][colName], columnInfo["type"].(string))
			if err != nil {
				return wrapError(tableName, colName, err)
			}
			err = checkCodes(columnInfo)
			if err != nil {
				return wrapError(tableName, colName, err)
			}
		}
	}
	return nil
}

// checks if the expected type string matches the received type
func checkExpectedType(expectedType string, receivedType string) error {
	// behavior should change with config, if lax, this should give a warning and coerce the received type to expected and log this transformation
	// if strict, return a genuine error
	if expectedType != receivedType {
		return strategy.UnexpectedTypeError{ExpectedType: expectedType, ActualType: receivedType}
	}
	return nil
}

// check if the codes for the column is correct
func checkCodes(columnInfo map[string]any) error {
	var s strategy.Strategy
	colType := utils.TrimAndUpperString(columnInfo["type"].(string)) // get the type as string
	code := utils.TrimAndUpperString(columnInfo["code"].(string))    // get the code as string
	val := columnInfo["value"]                                       // get the value
	err := preprocess(&val, colType)
	if err != nil {
		return err
	}
	s, err = strategy.GetStrategy(colType, code)
	if err != nil {
		return err
	}

	if _, ok := s.(strategy.ValueStrategy); ok {
		valStrategy := s.(strategy.ValueStrategy)
		valStrategy.SetValue(val)
		err = valStrategy.CheckCriteria()
		if err != nil {
			return err
		}
		return nil
	} else {
		err = s.CheckCriteria()
		if err != nil {
			return err
		}
		return nil
	}
}
