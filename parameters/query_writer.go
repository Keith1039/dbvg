package parameters

import (
	"container/list"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Keith1039/Capstone_Test/db"
	"github.com/Keith1039/Capstone_Test/graph"
	"log"
	"os"
	"strings"
)

func NewQueryWriterFor(db *sql.DB, tableName string) (*QueryWriter, error) {
	qw := QueryWriter{db: db, TableName: tableName} // set the table name
	err := qw.Init()                                // init the writer
	if err != nil {
		return nil, err // return the error
	}
	return &qw, nil // return the writer
}

func NewQueryWriterWithTemplateFor(db *sql.DB, tableName string, filePath string) (*QueryWriter, error) {
	// check to see if file exists
	m := make(map[string]map[string]map[string]string)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, err
	}
	qw := QueryWriter{db: db, TableName: tableName}
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
		// TODO make an error for this and give more detail to end user
		log.Fatal(err)
	}
	qw.updateTableMap(m) // actually update the table
	return &qw, nil      // return new parser
}

type QueryWriter struct {
	db               *sql.DB
	TableName        string
	tableMap         map[string]*table
	allRelations     map[string]map[string]map[string]string
	fkMap            map[string]map[string]string
	TableOrderQueue  *list.List // queue
	InsertQueryQueue *list.List // queue
	DeleteQueryQueue *list.List // queue
}

func (qw *QueryWriter) Init() error {
	var err error
	qw.TableName = strings.ToLower(qw.TableName)
	ordering := graph.NewOrdering(qw.db) // get a new ordering
	qw.allRelations = db.CreateRelationships(qw.db)
	qw.setFKMap()
	qw.TableOrderQueue, err = ordering.GetOrder(qw.TableName) // get the topological ordering of tables
	if err != nil {
		return err
	}
	qw.setTableMap()
	qw.InsertQueryQueue = list.New()
	qw.DeleteQueryQueue = list.New()
	return err
}

func (qw *QueryWriter) ChangeTableToWriteFor(tableName string) error {
	// change the table name of the writer and return any errors
	qw.TableName = strings.ToLower(tableName)
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

func (qw *QueryWriter) GenerateEntries(amount int) {
	for i := 0; i < amount; i++ {
		node := qw.TableOrderQueue.Front()
		for node != nil {
			qw.processTable(node.Value.(string))
			node = node.Next()
		}
	}
}

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
	// TODO have this return a custom error
	relations := db.CreateRelationships(qw.db)
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
