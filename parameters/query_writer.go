package parameters

import (
	"container/list"
	"fmt"
	"github.com/Keith1039/Capstone_Test/db"
	"github.com/Keith1039/Capstone_Test/graph"
	"log"
	"strings"
)

type QueryWriter struct {
	TableName        string
	AllRelations     map[string]map[string]map[string]string
	pkMap            map[string]string
	fkMap            map[string]map[string]string
	TableOrderQueue  *list.List // queue
	InsertQueryQueue *list.List // queue
	DeleteQueryQueue *list.List // queue
}

func (qw *QueryWriter) Init() error {
	var err error
	qw.TableName = strings.ToLower(qw.TableName)
	ordering := graph.Ordering{}
	ordering.Init()
	qw.AllRelations = db.CreateRelationships()
	qw.pkMap = db.GetTablePKMap()
	qw.SetFKMap()
	qw.TableOrderQueue, err = ordering.FindOrder(qw.TableName) // get the topological ordering of tables
	qw.InsertQueryQueue = list.New()
	qw.DeleteQueryQueue = list.New()
	return err
}

func (qw *QueryWriter) SetFKMap() {
	m := make(map[string]map[string]string)
	for _, relations := range qw.AllRelations {
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
	for qw.TableOrderQueue.Len() > 0 {
		qw.ProcessTable()
	}
}

func (qw *QueryWriter) ProcessTable() {
	//var writer SQLWriter
	colString := "("
	colValString := "("
	tableName := qw.TableOrderQueue.Front().Value.(string)
	t := createTable(tableName)
	for _, col := range t.Columns {
		fkRelation, fk := qw.AllRelations[tableName][col.ColumnName]
		if fk {
			colVal := qw.fkMap[fkRelation["Table"]][fkRelation["Column"]] // retrieve the stored foreign key value
			appendValues(&colString, &colValString, col.ColumnName, colVal)
		} else {
			colVal, err := col.Parser.ParseColumn()
			if err != nil {
				log.Fatal(err)
			}
			_, isFK := qw.fkMap[tableName][col.ColumnName]
			if isFK {
				qw.fkMap[tableName][col.ColumnName] = colVal
			}
			appendValues(&colString, &colValString, col.ColumnName, colVal)
		}
	}
	colString = colString + ")"
	colValString = colValString + ")"
	query := fmt.Sprintf("INSERT INTO %s %s VALUES %s;", t.TableName, colString, colValString)
	qw.InsertQueryQueue.PushBack(query)
	qw.TableOrderQueue.Remove(qw.TableOrderQueue.Front()) // remove the first in the queue
}

func createTable(tableName string) table {
	t := table{TableName: tableName}
	columnMap := db.GetColumnMap(tableName)
	columns := make([]column, len(columnMap))
	i := 0
	for columnName, dataType := range columnMap {
		c := column{ColumnName: columnName, Type: dataType, Parser: getColumnParser(dataType)}
		columns[i] = c
		i++
	}
	t.Columns = columns
	return t
}

func appendValues(colStringPtr *string, valStringPtr *string, newColumn string, newVal string) {
	if *colStringPtr == "(" {
		*colStringPtr = *colStringPtr + newColumn
	} else {
		*colStringPtr = *colStringPtr + "," + newColumn
	}

	if *valStringPtr == "(" {
		*valStringPtr = *valStringPtr + newVal
	} else {
		*valStringPtr = *valStringPtr + "," + newVal
	}
}
