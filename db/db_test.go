package db_test

import (
	"database/sql"
	"fmt"
	database "github.com/Keith1039/dbvg/db"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/golangmigrator"
	"maps"
	"slices"
	"testing"
	"time"
)

var db *sql.DB

var pgConf pgtestdb.Config

var migrator pgtestdb.Migrator

func init() {
	pgConf = pgtestdb.Config{
		DriverName: "postgres", // uses the lib/pq driver
		//Database:   "postgres",
		User:     "postgres",
		Password: "password",
		Host:     "localhost",
		Port:     "2000",
		Options:  "sslmode=disable",
	}
}

func TestInitAndCloseDB(t *testing.T) {
	// valid connection url
	db2, err := database.InitDB(pgConf.URL())
	if err != nil {
		t.Fatal(err)
	}
	defer database.CloseDB(db2)
	err = db2.Ping()
	if err != nil {
		t.Fatal(err)
	}
}

func TestIsSupportedType(t *testing.T) {
	ok := database.IsSupportedType("int")
	if ok {
		t.Fatal("lower case type name should be invalid")
	}
	ok = database.IsSupportedType("SOMERAndom")
	if ok {
		t.Fatal("unsupported type name should be invalid")
	}
	ok = database.IsSupportedType("VARCHAR")
	if !ok {
		t.Fatal("VARCHAR is a supported type but was deemed invalid")
	}
}

func TestGetTableMap(t *testing.T) {
	migrator = golangmigrator.New("migrations/case5")
	db = pgtestdb.New(t, pgConf, migrator)
	expectedMap := map[string]bool{
		"a": true,
		"b": true,
		"c": true,
		"d": true,
		"e": true,
		"f": true,
	}
	data := database.GetTableMap(db)
	delete(data, "schema_migrations")
	if !maps.Equal(expectedMap, data) {
		t.Fatalf("Expected: %v, Actual: %v", expectedMap, data)
	}
}

func TestGetColumnMap(t *testing.T) {
	migrator = golangmigrator.New("real_migrations/")
	db = pgtestdb.New(t, pgConf, migrator)
	expectedMap := map[string]string{
		"user_id":    "UUID",
		"product_id": "UUID",
		"quantity":   "INT",
		"created_at": "DATE",
	}
	data := database.GetColumnMap(db, "purchases")
	delete(data, "schema_migrations")
	if !maps.Equal(expectedMap, data) {
		t.Fatalf("Expected: %v \nActual: %v", expectedMap, data)
	}
}

func TestGetAllColumnData(t *testing.T) {
	migrator = golangmigrator.New("real_migrations/")
	db = pgtestdb.New(t, pgConf, migrator)
	expectedMap := map[string]map[string]string{
		"users": {
			"id":         "UUID",
			"first_name": "VARCHAR",
			"last_name":  "VARCHAR",
			"email":      "VARCHAR",
			"address":    "VARCHAR",
			"created_at": "DATE",
		},
		"companies": {
			"id":         "UUID",
			"name":       "VARCHAR",
			"email":      "VARCHAR",
			"created_at": "DATE",
		},
		"products": {
			"id":          "UUID",
			"company_id":  "UUID",
			"item_name":   "VARCHAR",
			"price":       "FLOAT",
			"quantity":    "INT",
			"description": "VARCHAR",
			"created_at":  "DATE",
		},
		"purchases": {
			"user_id":    "UUID",
			"product_id": "UUID",
			"quantity":   "INT",
			"created_at": "DATE",
		},
	}
	data := database.GetAllColumnData(db)
	delete(data, "schema_migrations")
	if len(data) != len(expectedMap) {
		t.Fatalf("Expected: %v, Actual: %v", expectedMap, data)
	}
	for key, value := range data {
		if !maps.Equal(expectedMap[key], value) {
			t.Fatalf("for key '%s',\nExpected: %v \nActual: %v", key, expectedMap[key], value)
		}
	}
}

func TestGetRawColumnMap(t *testing.T) {
	migrator = golangmigrator.New("real_migrations/")
	db = pgtestdb.New(t, pgConf, migrator)
	validateMap := map[string]string{
		"id":          "UUID",
		"company_id":  "UUID",
		"item_name":   "VARCHAR",
		"price":       "MONEY",
		"quantity":    "INT4",
		"description": "VARCHAR",
		"created_at":  "TIMESTAMP",
	}
	data := database.GetRawColumnMap(db, "products")
	delete(data, "schema_migrations")
	if len(data) != len(validateMap) {
		t.Fatalf("Expected: %v \nActual: %v", validateMap, data)
	}
	for key, value := range data {
		if validateMap[key] != value.DatabaseTypeName() {
			t.Fatalf("for key '%s',\nExpected: %v \nActual: %v", key, validateMap[key], value.DatabaseTypeName())
		}
	}
}

func TestGetTablePKMap(t *testing.T) {
	migrator = golangmigrator.New("migrations/case8/")
	db = pgtestdb.New(t, pgConf, migrator)
	expectedMap := map[string][]string{
		"a": {"akey"},
		"b": {"bkey", "bkey2"},
		"c": {"ckey"},
		"d": {"dkey"},
		"e": {"ekey"},
	}
	data := database.GetTablePKMap(db)
	delete(data, "schema_migrations")
	if len(data) != len(expectedMap) {
		t.Fatalf("Expected: %v \nActual: %v", expectedMap, data)
	}
	for key, value := range data {
		if !slices.Equal(expectedMap[key], value) {
			t.Fatalf("for key '%s'\nExpected: %v\nActual: %v", key, expectedMap[key], value)
		}
	}
}

func TestGetRelationships(t *testing.T) {
	migrator = golangmigrator.New("migrations/case9/")
	db = pgtestdb.New(t, pgConf, migrator)
	expectedMap := map[string]map[string]map[string]string{
		"a": {
			"bkey": {
				"Table":  "b",
				"Column": "bkey",
			},
		},
		"b": {
			"bkey_ref": {
				"Table":  "b",
				"Column": "bkey",
			},
			"ckey": {
				"Table":  "c",
				"Column": "ckey",
			},
			"dkey": {
				"Table":  "d",
				"Column": "dkey",
			},
			"ikey": {
				"Table":  "i",
				"Column": "ikey",
			},
		},
		"c": {
			"akey": {
				"Table":  "a",
				"Column": "akey",
			},
		},
		"d": {
			"ekey": {
				"Table":  "e",
				"Column": "ekey",
			},
		},
		"e": {
			"bkey": {
				"Table":  "b",
				"Column": "bkey",
			},
		},
		"f": {
			"gkey": {
				"Table":  "g",
				"Column": "gkey",
			},
			"bkey": {
				"Table":  "b",
				"Column": "bkey",
			},
		},
		"g": {
			"hkey": {
				"Table":  "h",
				"Column": "hkey",
			},
		},
		"h": {
			"fkey": {
				"Table":  "f",
				"Column": "fkey",
			},
		},
		"i": {
			"jkey": {
				"Table":  "j",
				"Column": "jkey",
			},
		},
		"j": {
			"kkey": {
				"Table":  "k",
				"Column": "kkey",
			},
		},
		"k": {
			"ikey": {
				"Table":  "i",
				"Column": "ikey",
			},
		},
		"l": {
			"mkey": {
				"Table":  "m",
				"Column": "mkey",
			},
		},
		"m": {
			"lkey": {
				"Table":  "l",
				"Column": "lkey",
			},
		},
	}
	data := database.GetRelationships(db)
	delete(data, "schema_migrations")
	if len(data) != len(expectedMap) {
		t.Fatalf("Expected: %v \nActual: %v", expectedMap, data)
	}
	for table, columnInfo := range data {
		for name, fkRelations := range columnInfo {
			if !maps.Equal(expectedMap[table][name], fkRelations) {
				t.Fatalf("for table '%s' and column '%s'\n Expected: %v\nActual: %v", table, name, expectedMap[table][name], fkRelations)
			}
		}
	}
}

func TestGetInverseRelationships(t *testing.T) {
	migrator = golangmigrator.New("migrations/case9/")
	db = pgtestdb.New(t, pgConf, migrator)
	expectedMap := map[string]map[string]bool{
		"a": {
			"c": true,
		},
		"b": {
			"a": true,
			"b": true,
			"e": true,
			"f": true,
		},
		"c": {
			"b": true,
		},
		"d": {
			"b": true,
		},
		"e": {
			"d": true,
		},
		"f": {
			"h": true,
		},
		"g": {
			"f": true,
		},
		"h": {
			"g": true,
		},
		"i": {
			"k": true,
			"b": true,
		},
		"j": {
			"i": true,
		},
		"k": {
			"j": true,
		},
		"l": {
			"m": true,
		},
		"m": {
			"l": true,
		},
	}
	data := database.GetInverseRelationships(db)
	delete(data, "schema_migrations")
	if len(data) != len(expectedMap) {
		t.Fatalf("Expected: %v \nActual: %v", expectedMap, data)
	}
	for key, value := range data {
		if !maps.Equal(expectedMap[key], value) {
			t.Fatalf("for key '%s'\nExpected: %v\nActual: %v", key, expectedMap[key], value)
		}
	}

}

func TestRunUnsafeQueries(t *testing.T) {
	migrator = golangmigrator.New("real_migrations/")
	db = pgtestdb.New(t, pgConf, migrator)
	queries := []string{
		fmt.Sprintf("INSERT INTO USERS(ID, FIRST_NAME, LAST_NAME, EMAIL, ADDRESS, CREATED_AT) VALUES ('%v', '%v', '%v', '%v', '%v', '%v')", uuid.New(), "some", "name", "test@gmail.com", "SOMETHING", time.Now().Format("2006-01-02 15:04:05")),
		fmt.Sprintf("INSERT INTO USERS(ID, FIRST_NAME, LAST_NAME, EMAIL, ADDRESS, CREATED_AT) VALUES ('%v', '%v', '%v', '%v', '%v', '%v')", uuid.New(), "some", "name", "test2@gmail.com", "SOMETHING", time.Now().Format("2006-01-02 15:04:05")),
	}
	err := database.RunUnsafeQueries(db, queries, false)
	if err != nil {
		t.Fatal(err)
	}
	err = checkCountForTable(db, "USERS", 2)
	if err != nil {
		t.Fatal(err)
	}
	// insert that violates unique
	queries = []string{
		fmt.Sprintf("INSERT INTO USERS(ID, FIRST_NAME, LAST_NAME, EMAIL, ADDRESS, CREATED_AT) VALUES ('%v', '%v', '%v', '%v', '%v', '%v')", uuid.New(), "some", "name", "test@gmail.com", "SOMETHING", time.Now().Format("2006-01-02 15:04:05")),
	}
	err = database.RunUnsafeQueries(db, queries, false)
	if err == nil {
		t.Fatal("query should violate emails unique constraint")
	}
	// check to see if roll back occurred
	err = checkCountForTable(db, "USERS", 2)
	if err != nil {
		t.Fatal(err)
	}

}

func checkCountForTable(db *sql.DB, tableName string, expected int) error {
	var count int
	row, err := db.Query(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName))
	if err != nil {
		return err
	}
	defer row.Close()
	row.Next()
	err = row.Scan(&count)
	if err != nil {
		return err
	}
	if count != expected {
		return fmt.Errorf("expected %d rows but got %d", expected, count)
	}
	return nil
}
