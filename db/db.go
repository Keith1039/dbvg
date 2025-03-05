// Package db is responsible for any database querying
//
// The db package provides a set of exported functions that returns formatted database outputs that are usable in the other packages
package db

import (
	"database/sql"
	"fmt"
	"github.com/jimsmart/schema"
	_ "github.com/lib/pq"
	"log"
)

const postgresFKRelations = `
	SELECT 
		KCU1.TABLE_NAME AS FK_TABLE_NAME, 
		KCU1.COLUMN_NAME AS FK_COLUMN_NAME, 
		KCU2.TABLE_NAME AS REFERENCED_TABLE_NAME, 
		KCU2.COLUMN_NAME AS REFERENCED_COLUMN_NAME
	FROM INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS AS RC 
	
	INNER JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE AS KCU1 
		ON KCU1.CONSTRAINT_CATALOG = RC.CONSTRAINT_CATALOG  
		AND KCU1.CONSTRAINT_SCHEMA = RC.CONSTRAINT_SCHEMA 
		AND KCU1.CONSTRAINT_NAME = RC.CONSTRAINT_NAME 
	
	INNER JOIN INFORMATION_SCHEMA.KEY_COLUMN_USAGE AS KCU2 
		ON KCU2.CONSTRAINT_CATALOG = RC.UNIQUE_CONSTRAINT_CATALOG  
		AND KCU2.CONSTRAINT_SCHEMA = RC.UNIQUE_CONSTRAINT_SCHEMA 
		AND KCU2.CONSTRAINT_NAME = RC.UNIQUE_CONSTRAINT_NAME 
		AND KCU2.ORDINAL_POSITION = KCU1.ORDINAL_POSITION 
`

var typeMap = map[string]string{
	"INT4":    "INT",
	"INT8":    "INT",
	"INT16":   "INT",
	"INT32":   "INT",
	"INT64":   "INT",
	"UUID":    "UUID",
	"VARCHAR": "VARCHAR",
	"BOOL":    "BOOL",
	"DATE":    "DATE",
}

// InitDB takes in a connection string and returns a database connection alongside any errors that occur
func InitDB(connString string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connString)
	return db, err
}

// CloseDB takes in a database connection and calls the Close() fun and logs any errors
func CloseDB(db *sql.DB) {
	err := db.Close()
	if err != nil {
		log.Fatal(err)
	}
}

// GetTableMap returns a map of existing table names mapped to the number 1 in the given database
func GetTableMap(db *sql.DB) map[string]int {
	tnames, err := schema.TableNames(db)
	allNames := make(map[string]int)
	if err != nil {
		log.Fatal(err)
	}
	for i := range tnames {
		tableName := tnames[i][1]
		allNames[tableName] = 1
	}
	return allNames
}

// GetColumnMap returns a map of column names mapped to their "translated" string type
func GetColumnMap(db *sql.DB, tableName string) map[string]string {
	m := make(map[string]string)                        // make initial map
	tcols, err := schema.ColumnTypes(db, "", tableName) // get the column info
	if err != nil {
		log.Fatal(err)
		return nil
	}
	for i := range tcols {
		m[tcols[i].Name()] = typeMap[tcols[i].DatabaseTypeName()] // map the column name to it's type
	}
	return m
}

// GetRawColumnMap returns a map of column names mapped to their string type
func GetRawColumnMap(db *sql.DB, tableName string) map[string]string {
	m := make(map[string]string)                        // make initial map
	tcols, err := schema.ColumnTypes(db, "", tableName) // get the column info
	if err != nil {
		log.Fatal(err)
		return nil
	}
	for i := range tcols {
		m[tcols[i].Name()] = tcols[i].DatabaseTypeName() // map the column name to it's type
	}
	return m
}

// GetTablePKMap returns a map of table names mapped to an array of string primary keys
func GetTablePKMap(db *sql.DB) map[string][]string {
	var pks []string
	tnames, err := schema.TableNames(db)
	pkMap := make(map[string][]string)
	if err != nil {
		log.Fatal(err)
	}
	for i := range tnames {
		tableName := tnames[i][1]
		pks, err = schema.PrimaryKey(db, "", tableName)
		if err != nil {
			log.Fatal(err)
		}
		pkMap[tableName] = pks
	}
	return pkMap
}

// GetRelationships returns a map relating tables to each other via their columns. Format: {table1: {FKColumn: {"Table: table2, "Column": "table2_Col"}}}
func GetRelationships(db *sql.DB) map[string]map[string]map[string]string {
	relations := make(map[string]map[string]map[string]string)
	var tableName, fkColumnName, refTableName, refColumnName string
	rows, err := db.Query(postgresFKRelations)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		err = rows.Scan(&tableName, &fkColumnName, &refTableName, &refColumnName)
		if err != nil {
			log.Fatal(err)
		}
		table, ok := relations[tableName]
		if !ok {
			relations[tableName] = map[string]map[string]string{fkColumnName: {"Table": refTableName, "Column": refColumnName}}
		} else {
			table[fkColumnName] = map[string]string{"Table": refTableName, "Column": refColumnName}
		}
	}
	return relations
}

// GetInverseRelationships returns a map relating tables to tables that relate to them. It's the same data as the `GetRelationship()` map
// but formatted differently
func GetInverseRelationships(db *sql.DB) map[string]map[string]bool {
	var tableName, fkColumnName, refTableName, refColumnName string
	relations := make(map[string]map[string]bool)
	rows, err := db.Query(postgresFKRelations)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		err = rows.Scan(&tableName, &fkColumnName, &refTableName, &refColumnName)
		if err != nil {
			log.Fatal(err)
		}
		table, ok := relations[refTableName]
		if !ok {
			relations[refTableName] = map[string]bool{tableName: true}
		} else {
			table[tableName] = true
		}
	}
	return relations
}

// RunQueries runs a given list of queries and returns any errors the moment they happen
func RunQueries(db *sql.DB, queries []string) error {
	var err error
	for _, query := range queries {
		_, err = db.Exec(query)
		if err != nil {
			return err
		}
	}
	return err
}

// RunQueriesVerbose runs a given list of queries but prints them out before executing them
func RunQueriesVerbose(db *sql.DB, queries []string) error {
	var err error

	for i, query := range queries {

		fmt.Println(fmt.Sprintf("Query %d: %s", i+1, query))
		_, err = db.Exec(query)
		if err != nil {
			return err
		}
	}
	return err
}
