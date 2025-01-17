package db

import (
	"database/sql"
	"fmt"
	"github.com/jimsmart/schema"
	_ "github.com/lib/pq"
	"log"
	"os"
)

var db *sql.DB

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
	"VARCHAR": "VARCHAR",
	"BOOL":    "BOOL",
	"DATE":    "DATE",
}

func init() {
	var err error
	err = os.Setenv("DATABASE_URL", "postgres://postgres:localDB12@localhost:5432/testgres?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
}

func DisplayTable() {
	// Fetch names of all tables
	tnames, err := schema.TableNames(db)
	if err != nil {
		log.Fatal(err)
	}
	// tnames is [][2]string
	for i := range tnames {
		tableName := tnames[i][1]
		fmt.Println("Table:", tableName)
		// Fetch column metadata for given table
		tcols, _ := schema.ColumnTypes(db, "", tableName)
		// tcols is []*sql.ColumnType
		for i := range tcols {
			fmt.Println("Column:", tcols[i].Name(), tcols[i].DatabaseTypeName())
		}
		// Fetch primary key for given table
		pks, _ := schema.PrimaryKey(db, "", tableName)

		// pks is []string
		for i := range pks {
			fmt.Println("Primary Key:", pks[i])
		}
		fmt.Println("........................")
	}
}

func GetTableMap() map[string]int {
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

func GetColumnMap(tableName string) map[string]string {
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

func GetRawColumnMap(tableName string) map[string]string {
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

func GetTablePKMap() map[string]string {
	tnames, err := schema.TableNames(db)
	pkMap := make(map[string]string)
	if err != nil {
		log.Fatal(err)
	}
	for i := range tnames {
		tableName := tnames[i][1]
		pks, err := schema.PrimaryKey(db, "", tableName)
		if err != nil {
			log.Fatal(err)
		}
		if len(pks) >= 1 {
			pkMap[tableName] = pks[0]
		} else {
			pkMap[tableName] = ""
		}
	}
	return pkMap
}

func CreateRelationships() map[string]map[string]map[string]string {
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

func CreateInverseRelationships() map[string]map[string]map[string]string {
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
		table, ok := relations[refTableName]
		if !ok {
			relations[refTableName] = map[string]map[string]string{tableName: {"Column": refColumnName, "FKColumn": fkColumnName}}
		} else {
			table[tableName] = map[string]string{"Column": refColumnName, "FKColumn": fkColumnName}
		}
	}
	return relations
}

func CreateRelationshipsWithDB(database *sql.DB) map[string]map[string]map[string]string {
	relations := make(map[string]map[string]map[string]string)
	var tableName, fkColumnName, refTableName, refColumnName string
	rows, err := database.Query(postgresFKRelations)
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
