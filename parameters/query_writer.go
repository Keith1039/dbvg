// Package parameters parses through parameters to generate data for a DB table
//
// the parameters package uses templates, either a default or a user defined template, to parse and generate data for a given database table
package parameters

import (
	"container/list"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/graph"
	"log"
	"os"
	"strings"
)

// NewQueryWriterFor takes in a database connection alongside a table name and returns a pointer to a QueryWriter that is initialized and any errors that occurred
func NewQueryWriterFor(db *sql.DB, tableName string) (*QueryWriter, error) {
	qw := QueryWriter{db: db, tableName: tableName} // set the table name
	err := qw.Init()                                // init the writer
	if err != nil {
		return nil, err // return the error
	}
	return &qw, nil // return the writer
}

// NewQueryWriterWithTemplateFor takes in a database connection, table name as well as a file path to a template that is used to set the values in the QueryWriter
// before returning a pointer to the initialized QueryWriter as well as any errors that occurred
func NewQueryWriterWithTemplateFor(db *sql.DB, tableName string, filePath string) (*QueryWriter, error) {
	// check to see if file exists
	m := make(map[string]map[string]map[string]string)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, err
	}
	qw := QueryWriter{db: db, tableName: tableName}
	err := qw.Init()
	if err != nil {
		return nil, err
	}

	bytes, err := os.ReadFile(filePath) // read the bytes
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &m) // unmarshal the JSON
	if err != nil {
		return nil, err
	}

	err = qw.verifyTemplate(m) // do further verifications on the template
	if err != nil {
		// print some error message
		log.Fatal(err)
	}
	qw.updateTableMap(m) // actually update the table
	return &qw, nil      // return new parser
}

// QueryWriter is the struct responsible for generating data for it's given table
type QueryWriter struct {
	db               *sql.DB                                 // the database connection
	tableName        string                                  // name of the 'table' it's generating data for
	tableMap         map[string]*table                       // a map of the 'table names' to their table object
	allRelations     map[string]map[string]map[string]string // all 'table' relationships expressed as a map
	fkMap            map[string]map[string]string            // a map of foreign keys
	TableOrderQueue  *list.List                              // queue
	InsertQueryQueue *list.List                              // queue
	DeleteQueryQueue *list.List                              // queue
}

// Init initializes the QueryWriter and returns any errors that occur upon initialization
func (qw *QueryWriter) Init() error {
	var err error
	qw.tableName = strings.ToLower(qw.tableName)
	ordering := graph.NewOrdering(qw.db) // get a new ordering
	qw.allRelations = db.GetRelationships(qw.db)
	qw.setFKMap()
	qw.TableOrderQueue, err = ordering.GetOrder(qw.tableName) // get the topological ordering of tables
	if err != nil {
		return err
	}
	qw.setTableMap()
	qw.InsertQueryQueue = list.New()
	qw.DeleteQueryQueue = list.New()
	return err
}

// ChangeTableToWriteFor takes in a new table name and re-inits the QueryWriter. It returns any errors that happen upon re-initialization
func (qw *QueryWriter) ChangeTableToWriteFor(tableName string) error {
	// change the table name of the writer and return any errors
	qw.tableName = strings.ToLower(tableName)
	return qw.Init()
}

func (qw *QueryWriter) setFKMap() {
	m := make(map[string]map[string]string)
	for _, relations := range qw.allRelations {
		for _, relation := range relations {
			r, ok := m[relation["Table"]]
			if !ok {
				m[relation["Table"]] = map[string]string{relation["Column"]: ""}
			} else {
				r[relation["Column"]] = ""
			}
		}
	}
	qw.fkMap = m
}

// GenerateEntries takes in a number and generates that amount of entries for the QueryWriter table
func (qw *QueryWriter) GenerateEntries(amount int) {
	for i := 0; i < amount; i++ {
		node := qw.TableOrderQueue.Front()
		for node != nil {
			qw.processTable(node.Value.(string))
			node = node.Next()
		}
	}
}

// GenerateEntry is a wrapper around the GenerateEntries function that simply gives the later an amount of 1
func (qw *QueryWriter) GenerateEntry() {
	qw.GenerateEntries(1) // only generate one
}

func (qw *QueryWriter) processTable(tableName string) {
	//var writer SQLWriter
	var colBuilder, colValBuilder, deleteBuilder strings.Builder
	colBuilder.WriteString("(")
	colValBuilder.WriteString("(")
	t := qw.tableMap[tableName]
	for _, col := range t.Columns {
		fkRelation, fk := qw.allRelations[tableName][col.ColumnName]
		if fk {
			colVal := qw.fkMap[fkRelation["Table"]][fkRelation["Column"]] // retrieve the stored foreign key value
			appendValues(&colBuilder, &colValBuilder, col.ColumnName, colVal)
			buildDeleteQuery(&deleteBuilder, col.ColumnName, colVal)
		} else {
			colVal, err := col.Parser.ParseColumn(*col)
			if err != nil {
				log.Fatal(err)
			}
			_, isFK := qw.fkMap[tableName][col.ColumnName]
			if isFK {
				qw.fkMap[tableName][col.ColumnName] = colVal
			}
			appendValues(&colBuilder, &colValBuilder, col.ColumnName, colVal)
			buildDeleteQuery(&deleteBuilder, col.ColumnName, colVal)
		}
	}
	colBuilder.WriteString(")")
	colValBuilder.WriteString(")")
	query := fmt.Sprintf("INSERT INTO %s %s VALUES %s;", t.TableName, colBuilder.String(), colValBuilder.String())
	qw.InsertQueryQueue.PushBack(query)
	qw.DeleteQueryQueue.PushFront(fmt.Sprintf("DELETE FROM %s WHERE %s;", tableName, deleteBuilder.String()))
}

func (qw *QueryWriter) setTableMap() {
	m := make(map[string]*table)
	node := qw.TableOrderQueue.Front()
	for node != nil {
		tableName := node.Value.(string)         // get the table
		tableStruct := qw.createTable(tableName) // create the table struct
		m[tableName] = &tableStruct              // map it
		node = node.Next()
	}
	qw.tableMap = m
}

func (qw *QueryWriter) createTable(tableName string) table {
	t := table{TableName: tableName}
	columnMap := db.GetColumnMap(qw.db, tableName)
	columns := make([]*column, len(columnMap))
	i := 0
	for columnName, dataType := range columnMap {
		parser := getColumnParser(dataType)
		c := column{ColumnName: columnName, Type: dataType, Parser: parser}
		columns[i] = &c
		i++
	}
	t.Columns = columns
	return t
}

func (qw *QueryWriter) verifyTemplate(m map[string]map[string]map[string]string) error {
	relations := db.GetRelationships(qw.db)
	flag := qw.TableOrderQueue.Len() == len(m) // number of keys should match number of tables
	if !flag {
		return errors.New("number of tables in template does not match the number of tables required")
	}
	// loop through the keys in the template
	for key := range m {
		t := qw.tableMap[key]
		// check if all the column names match
		for _, col := range t.Columns {
			_, exists := m[key][col.ColumnName] // check if there's an entry for the column in the template
			if !exists {
				_, exists = relations[key][col.ColumnName] // check to see if the column is missing because it's an FK
				if !exists {
					return errors.New(fmt.Sprintf("column %s from table %s exists in table but is missing in template and is not a foreign key reference", t.TableName, col.ColumnName))
				}
			}

			_, exists = stringToEnum[strings.ToUpper(m[key][col.ColumnName]["Code"])] // check to see if the code exists
			// check to see if the code doesn't exist AND the code isn't empty string
			if !exists && strings.TrimSpace(m[key][col.ColumnName]["Code"]) != "" {
				return errors.New(fmt.Sprintf("Code %s is not supported or recognized by parser of type %s", m[key][col.ColumnName]["Code"], col.Type))
			}

		}
	}
	return nil
}

func (qw *QueryWriter) updateTableMap(m map[string]map[string]map[string]string) {
	for key := range m {
		t := qw.tableMap[key]
		for _, col := range t.Columns {
			col.Code = stringToEnum[strings.ToUpper(m[key][col.ColumnName]["Code"])]
			col.Other = m[key][col.ColumnName]["Value"]
		}
	}
}

func appendValues(colBuilder *strings.Builder, colValBuilder *strings.Builder, newColumn string, newVal string) {
	if colBuilder.String() == "(" {
		colBuilder.WriteString(newColumn)
	} else {
		colBuilder.WriteString(fmt.Sprintf(", %s", newColumn))
	}

	if colValBuilder.String() == "(" {
		colValBuilder.WriteString(newVal)
	} else {
		colValBuilder.WriteString(fmt.Sprintf(", %s", newVal))
	}
}

func buildDeleteQuery(builder *strings.Builder, col string, val string) {
	if builder.Len() == 0 {
		builder.WriteString(fmt.Sprintf("%s=%s", col, val))
	} else {
		builder.WriteString(fmt.Sprintf(" AND %s=%s", col, val))
	}
}
