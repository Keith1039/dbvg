package parameters

import (
	"container/list"
	"fmt"
	"github.com/Keith1039/Capstone_Test/db"
	"github.com/Keith1039/Capstone_Test/graph"
	"log"
	"sort"
)

type QueryWriter struct {
	TableName        string
	AllRelations     map[string]map[string]map[string]string
	LevelMap         map[string]int
	pkMap            map[string]string
	fkMap            map[string]map[string]string
	TableOrderQueue  *list.List // queue
	InsertQueryQueue *list.List // queue
	DeleteQueryQueue *list.List // queue
}

func (qw *QueryWriter) Init() error {
	var err error
	ordering := graph.Ordering{}
	ordering.Init()
	qw.AllRelations = db.CreateRelationships()
	qw.LevelMap, err = ordering.FindOrder(qw.TableName)
	if err != nil {
		return err
	}
	qw.pkMap = db.GetTablePKMap()
	qw.SetFKMap()
	qw.TableOrderQueue = list.New()
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

func (qw *QueryWriter) CreateTableOrder() {
	l := list.New()
	tnames := make([]string, 0, len(qw.LevelMap))
	for key := range qw.LevelMap {
		tnames = append(tnames, key)
	}
	// sort in descending order
	sort.SliceStable(tnames, func(i, j int) bool {
		return qw.LevelMap[tnames[i]] > qw.LevelMap[tnames[j]]
	})
	for _, tname := range tnames {
		l.PushBack(tname) // push to the back of the queue
	}
	qw.TableOrderQueue = l // set the queue
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
