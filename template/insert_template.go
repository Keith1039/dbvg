package template

import (
	"github.com/Keith1039/dbvg/graph"
	"github.com/Keith1039/dbvg/strategy"
	"github.com/Keith1039/dbvg/utils"
	"maps"
	"slices"
	"strings"
)

var insertSchema = map[string]string{
	"code":  "string",
	"type":  "string",
	"value": "any",
}

func makeInsertTemplate(tableData map[string]map[string]string, requiredTables []string, path string) (*InsertTemplate, error) {
	t := &InsertTemplate{}
	t.strategyMap = make(map[string]map[string]StrategyCodePair) // initialize map
	err := t.templateFrom(tableData, requiredTables, path)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// NewInsertTemplateWithMap creates a InsertTable struct using a static map, this is to prevent multiple duplicate
// database queries. Process is identical to NewInsertTemplateWithDB
func NewInsertTemplateWithMap(tableData map[string]map[string]string, requiredTables []string, path string) (*InsertTemplate, error) {
	return makeInsertTemplate(tableData, requiredTables, path)
}

// StrategyCodePair is a struct only meant to facilitate the transfer of information between packages
type StrategyCodePair struct {
	Code     string            // exists solely for debugging
	Strategy strategy.Strategy // the thing we actually care about
}

func (s StrategyCodePair) IsEmpty() bool {
	return s.Strategy == nil && strings.TrimSpace(s.Code) == ""
}

// InsertTemplate is a struct that handles interactions with an Insert JSON template.
// to be more specific it maps tables and columns to their strategy by reading through JSON
type InsertTemplate struct {
	strategyMap map[string]map[string]StrategyCodePair
}

// fills the struct with data
func (t *InsertTemplate) templateFrom(tableData map[string]map[string]string, requiredTables []string, path string) error {
	data, err := retrieveJSON(path) // run some validation and retrieve JSON data
	if err != nil {                 // error check
		return err
	}
	err = t.validateTemplate(data, tableData, insertSchema, requiredTables) // validate the data given to see if it matches schema
	// add an extra step for checking if the value has a type that is supported by the code
	if err != nil { // error check
		return err
	}
	return nil
}

// validateTemplate confirms that the information gained from a JSON file is valid through a series of checks
func (t *InsertTemplate) validateTemplate(jsonData map[string]map[string]map[string]any, tableData map[string]map[string]string, schema map[string]string, requiredTables []string) error {
	var typeMap map[string]string

	requiredTableMap := make(map[string]bool)
	// check if names in required tables exist in map and add to map
	for _, tableName := range requiredTables {
		if _, ok := tableData[tableName]; !ok {
			return MissingRequiredTableError{tableName: tableName, jsonKeys: slices.Collect(maps.Keys(tableData))}
		} else {
			requiredTableMap[tableName] = true
		}
	}

	for tableName, columns := range jsonData { // loop through each key
		if _, ok := tableData[tableName]; !ok { // check if the tableName exists in the schema
			return graph.NewMissingTableError(tableName)
		}
		// condition to ignore irrelevant tables (i.e. any table that isn't required)
		if _, ok := requiredTableMap[tableName]; ok {
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
				err = t.checkCodesAndSetStrategy(tableName, colName, columnInfo)
				if err != nil {
					return wrapError(tableName, colName, err)
				}
			}
		}
	}
	return nil
}

// check if the codes for the column is correct
func (t *InsertTemplate) checkCodesAndSetStrategy(tableName, columnName string, columnInfo map[string]any) error {
	var s strategy.Strategy
	colType := utils.TrimAndUpperString(columnInfo["type"].(string)) // get the type as string
	code := utils.TrimAndUpperString(columnInfo["code"].(string))    // get the code as string
	val := columnInfo["value"]                                       // get the value
	if code == "" && val == nil {
		sFunc, ok := defaults[colType]
		if !ok {
			return UndefinedDefaultError{columnType: colType}
		}
		s = sFunc() // define s here
	} else {
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
		} else {
			err = s.CheckCriteria()
			if err != nil {
				return err
			}
		}
	}
	pair := StrategyCodePair{Code: code, Strategy: s}
	_, ok := t.strategyMap[tableName]
	if ok {
		t.strategyMap[tableName][columnName] = pair
	} else {
		t.strategyMap[tableName] = map[string]StrategyCodePair{columnName: pair}
	}
	return nil
}

func (t *InsertTemplate) GetStrategyCodePair(tableName, columnName string) StrategyCodePair {
	tableName = utils.TrimAndLowerString(tableName)
	columnName = utils.TrimAndLowerString(columnName)
	return t.strategyMap[tableName][columnName]
}
