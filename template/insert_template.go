package template

import (
	"database/sql"
	database "github.com/Keith1039/dbvg/db"
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

func makeInsertTemplate(db *sql.DB, table string, path string) (*InsertTemplate, error) {
	t := &InsertTemplate{}
	t.strategyMap = make(map[string]map[string]StrategyCodePair) // initialize map
	err := t.templateFrom(db, table, path)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// NewInsertTemplate creates a InsertTemplate for a table from an existing template indicated by
// the path variable
func NewInsertTemplate(db *sql.DB, table string, path string) (*InsertTemplate, error) {
	if strings.TrimSpace(path) == "" {
		return nil, MissingPathError{}
	}
	return makeInsertTemplate(db, table, path)
}

// NewDefaultInsertTemplate creates a InsertTemplate struct with a standard template
func NewDefaultInsertTemplate(db *sql.DB, table string) (*InsertTemplate, error) {
	return makeInsertTemplate(db, table, "")
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
func (t *InsertTemplate) templateFrom(db *sql.DB, table string, path string) error {
	var data map[string]map[string]map[string]any
	var err error
	var ord *graph.Ordering
	var order []string
	if path != "" {
		data, err = utils.RetrieveInsertTemplateJSON(path) // run some validation and retrieve JSON data
		if err != nil {                                    // error check
			return err
		}
	} else {
		ord, err = graph.NewOrdering(db)
		if err != nil {
			return err
		}
		order, err = ord.GetOrder(table)
		if err != nil {
			return err
		}
		data = utils.MakeTemplates(db, order)
	}
	err = t.validateTemplate(db, table, data, insertSchema, path == "") // validate the data given to see if it matches schema
	// add an extra step for checking if the value has a type that is supported by the code
	if err != nil { // error check
		return err
	}
	return nil
}

// validateTemplate confirms that the information gained from a JSON file is valid through a series of checks
func (t *InsertTemplate) validateTemplate(db *sql.DB, table string, jsonData map[string]map[string]map[string]any, schema map[string]string, validated bool) error {
	var typeMap map[string]string
	if !validated {
		ord, err := graph.NewOrdering(db)
		if err != nil {
			return err
		}
		tableData := database.GetAllColumnData(db)
		allRelations := database.GetRelationships(db)
		requiredTables, err := ord.GetOrder(table)
		if err != nil {
			return err
		}
		requiredTableMap := make(map[string]bool)
		requiredColMap := make(map[string]bool)
		// check if names in required tables exist in template
		for _, tableName := range requiredTables {
			if _, ok := jsonData[tableName]; !ok {
				return MissingRequiredTableError{tableName: tableName, jsonKeys: slices.Collect(maps.Keys(tableData))}
			} else {
				// check if we have all the right columns for the table (all non-FK tables)
				for col := range tableData[tableName] {
					_, ok = jsonData[tableName][col]
					_, isFk := allRelations[tableName][col]
					if !ok && !isFk {
						return MissingRequiredColumnError{tableName: tableName, columnName: col}
					} else if ok && !isFk {
						requiredColMap[col] = true
					}
				}
				requiredTableMap[tableName] = true
			}
		}

		for tableName, columns := range jsonData { // loop through each key
			// condition to ignore irrelevant tables (i.e. any table that isn't required)
			if _, ok := requiredTableMap[tableName]; ok {
				for colName, columnInfo := range columns {
					// ignore irrelevant columns
					if _, ok = requiredColMap[colName]; ok {
						normalizeKeys(columnInfo)                 // trims and lowers each key while maintaining the key value pairs
						typeMap = makeTypeMap(columnInfo)         // check the types for the key value pairs in the template
						err = checkAgainstSchema(typeMap, schema) // check the type map against the schema
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
		}
	} else {
		for tableName, columns := range jsonData { // loop through each key
			for colName, columnInfo := range columns {
				err := t.checkCodesAndSetStrategy(tableName, colName, columnInfo)
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
		code = defaultCode[colType] // set code
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
