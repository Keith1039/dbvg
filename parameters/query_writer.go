// Package parameters parses through parameters to generate data for a DB table
//
// the parameters package uses templates, either a default or a user defined template, to parse and generate data for a given database table
package parameters

import (
	"database/sql"
	"fmt"
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/graph"
	"github.com/Keith1039/dbvg/strategy"
	"github.com/Keith1039/dbvg/template"
	"github.com/Keith1039/dbvg/utils"
	"log"
	"os"
	"strings"
)

// NewQueryWriter takes in a database connection alongside a table name and returns a pointer to a QueryWriter that is initialized and any errors that occurred
func NewQueryWriter(db *sql.DB, tableName string) (*QueryWriter, error) {
	qw := QueryWriter{db: db, tableName: tableName} // set the table name
	err := qw.init()                                // init the writer
	if err != nil {
		return nil, err // return the error
	}
	qw.template, err = template.NewDefaultInsertTemplate(db, qw.TableOrder)
	if err != nil {
		return nil, err
	}
	qw.setTableMap()
	return &qw, nil // return the writer
}

// NewQueryWriterWithTemplate takes in a database connection, table name as well as a file path to a template that is used to set the values in the QueryWriter
// before returning a pointer to the initialized QueryWriter as well as any errors that occurred
func NewQueryWriterWithTemplate(db *sql.DB, tableName string, filePath string) (*QueryWriter, error) {
	// check to see if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, err
	}
	qw := QueryWriter{db: db, tableName: tableName}
	err := qw.init()
	if err != nil {
		return nil, err
	}
	qw.template, err = template.NewInsertTemplate(database.GetAllColumnData(db), qw.TableOrder, filePath)
	if err != nil {
		return nil, err
	}
	qw.setTableMap()
	return &qw, nil // return new parser
}

// QueryWriter is the struct responsible for generating data for it's given table
type QueryWriter struct {
	db           *sql.DB                                 // the database connection
	template     *template.InsertTemplate                // the template struct that contains the pairs
	tableName    string                                  // name of the 'table' it's generating data for
	tableMap     map[string]*table                       // a map of the 'table names' to their table object
	allRelations map[string]map[string]map[string]string // all 'table' relationships expressed as a map
	fkMap        map[string]map[string]any               // a map of foreign keys
	TableOrder   []string                                // queue
}

// init initializes the QueryWriter and returns any errors that occur upon initialization
func (qw *QueryWriter) init() error {
	qw.tableName = utils.TrimAndLowerString(qw.tableName)
	ordering, err := graph.NewOrdering(qw.db) // get a new ordering
	if err != nil {
		return err
	}
	qw.allRelations = database.GetRelationships(qw.db)
	qw.setFKMap()
	qw.TableOrder, err = ordering.GetOrder(qw.tableName) // get the topological ordering of tables
	if err != nil {
		return err
	}
	return err
}

func (qw *QueryWriter) setFKMap() {
	m := make(map[string]map[string]any)
	for _, relations := range qw.allRelations {
		for _, relation := range relations {
			r, ok := m[relation["Table"]]
			if !ok {
				m[relation["Table"]] = map[string]any{relation["Column"]: ""}
			} else {
				r[relation["Column"]] = ""
			}
		}
	}
	qw.fkMap = m
}

// GenerateEntries takes in a number and generates that amount of entries in the form of INSERT and DELETE queries which it returns as a string array
func (qw *QueryWriter) GenerateEntries(amount int) (*InsertBatch, *DeleteBatch) {
	insertBatch := &InsertBatch{}
	deleteBatch := &DeleteBatch{}
	total := len(qw.tableMap) * amount // expected total
	insertBatch.init(total)
	deleteBatch.init(total)

	for i := 0; i < amount; i++ {
		for _, tableName := range qw.TableOrder {
			qw.processTable(tableName, insertBatch, deleteBatch)
		}
	}

	return insertBatch, deleteBatch
}

// GenerateEntry is a wrapper around the GenerateEntries function that simply gives the later an amount of 1
func (qw *QueryWriter) GenerateEntry() (*InsertBatch, *DeleteBatch) {
	return qw.GenerateEntries(1) // only generate one
}

func (qw *QueryWriter) processTable(tableName string, insertBatch *InsertBatch, deleteBatch *DeleteBatch) {
	var colVal any
	var err error
	t := qw.tableMap[tableName]
	allColumns, paraString, deleteQuery := getQueryParams(t.Columns)
	i := 0
	parameterArr := make([]any, len(allColumns))
	for _, col := range t.Columns {
		fkRelation, fk := qw.allRelations[tableName][col.ColumnName]
		if fk {
			colVal = qw.fkMap[fkRelation["Table"]][fkRelation["Column"]] // retrieve the stored foreign key value
		} else {
			colVal, err = col.Pair.Strategy.ExecuteStrategy()
			if err != nil {
				log.Fatalf("strategy execution for code '%s' failed for column '%s' of table '%s'", col.Pair.Code, col.ColumnName, tableName)
			}
			_, isFK := qw.fkMap[tableName][col.ColumnName]
			if isFK {
				qw.fkMap[tableName][col.ColumnName] = colVal
			}
		}
		parameterArr[i] = colVal
		i++
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);", t.TableName, strings.Join(allColumns, ", "), strings.Join(paraString, ", "))
	insertBatch.append(query, parameterArr)

	query = fmt.Sprintf("DELETE FROM %s WHERE %s;", tableName, strings.Join(deleteQuery, " AND "))
	deleteBatch.reverseAppend(query, parameterArr)
}

func (qw *QueryWriter) setTableMap() {
	m := make(map[string]*table)
	for _, tableName := range qw.TableOrder {
		tableStruct := qw.createTable(tableName) // create the table struct
		m[tableName] = &tableStruct              // map it
	}
	qw.tableMap = m
}

func (qw *QueryWriter) createTable(tableName string) table {
	t := table{TableName: tableName}
	columnMap := database.GetColumnMap(qw.db, tableName)
	columns := make([]*column, len(columnMap))
	columnDetailsMap := database.GetRawColumnMap(qw.db, tableName) // get the column details map to add details to the column struct
	i := 0
	for columnName, dataType := range columnMap {
		pair := qw.template.GetStrategyCodePair(tableName, columnName)
		colDetails := columnDetailsMap[columnName]
		// exception for regex
		if dataType == "VARCHAR" && pair.Code == "REGEX" {
			p2, ok := pair.Strategy.(*strategy.RequiredStrategy)
			length, _ := colDetails.Length()    // no point in checking since all types mapped to varchar give a valid length
			cond := length != -5 && length < 10 // check if it's not default VARCHAR or BPCHAR as well as the length being lower than 10
			if ok && p2.Value == template.DEFAULTREGEX && cond {
				p2.SetValue(fmt.Sprintf("[a-zA-Z]{%d}", length)) // change the default expression to fit the container
			}
		}
		c := column{ColumnName: columnName, ColumnDetails: columnDetailsMap[columnName], Type: dataType, Pair: pair}
		columns[i] = &c
		i++
	}
	t.Columns = columns
	return t
}
