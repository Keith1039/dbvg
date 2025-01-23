package parameters

import (
	"container/list"
	"database/sql"
	"fmt"
	"github.com/Keith1039/Capstone_Test/db"
	"github.com/Keith1039/Capstone_Test/graph"
	"log"
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

func (qw *QueryWriter) ProcessTables() {
	node := qw.TableOrderQueue.Front()
	for node != nil {
		qw.processTable(node.Value.(string))
		node = node.Next()
	}
	//for qw.TableOrderQueue.Len() > 0 {
	//	qw.processTable()
	//}
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
			colVal, err := col.Parser.ParseColumn()
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
	columns := make([]column, len(columnMap))
	i := 0
	for columnName, dataType := range columnMap {
		parser := getColumnParser(dataType)
		c := column{ColumnName: columnName, Type: dataType, Parser: parser}
		columns[i] = c
		i++
	}
	t.Columns = columns
	return t
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
