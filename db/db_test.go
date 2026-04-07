package db_test

import (
	"database/sql"
	"fmt"
	database "github.com/Keith1039/dbvg/db"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"log"
	"maps"
	"os"
	"slices"
	"testing"
	"time"
)

const path = "file://../db/migrations/"

const realMigrationPath = "file://../db/real_migrations/"

var db *sql.DB

func drop() {
	// drop the database
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}
	m, err2 := migrate.NewWithDatabaseInstance(
		path+"case1",
		"postgres", driver)
	if m != nil {
		err = m.Drop()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err2)
	}
}

func buildUp(caseName string) error {
	// migrate the schema up
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	m, err2 := migrate.NewWithDatabaseInstance(
		path+caseName,
		"postgres", driver)
	if m != nil {
		err = m.Up()
		if err != nil {
			return err
		}
	} else {
		return err2
	}
	return nil
}

func buildUpRealCase() error {
	// migrate the schema up
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	m, err2 := migrate.NewWithDatabaseInstance(
		realMigrationPath,
		"postgres", driver)
	if m != nil {
		err = m.Up()
		if err != nil {
			return err
		}
	} else {
		return err2
	}
	return nil
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
	drop()
}

func TestGetTableMap(t *testing.T) {
	drop()
	err := buildUp("case5")
	if err != nil {
		t.Fatal(err)
	}
	expectedMap := map[string]int{
		"a": 1,
		"b": 1,
		"c": 1,
		"d": 1,
		"e": 1,
		"f": 1,
	}
	data := database.GetTableMap(db)
	delete(data, "schema_migrations")
	if !maps.Equal(expectedMap, data) {
		t.Fatalf("Expected: %v, Actual: %v", expectedMap, data)
	}
}

func TestGetColumnMap(t *testing.T) {
	drop()
	err := buildUpRealCase()
	if err != nil {
		t.Fatal(err)
	}
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
	drop()
	err := buildUpRealCase()
	if err != nil {
		t.Fatal(err)
	}
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
	drop()
	err := buildUpRealCase()
	if err != nil {
		t.Fatal(err)
	}
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
	drop()
	err := buildUp("case8")
	if err != nil {
		t.Fatal(err)
	}
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
	drop()
	err := buildUp("case9")
	if err != nil {
		t.Fatal(err)
	}
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
	drop()
	err := buildUp("case9")
	if err != nil {
		t.Fatal(err)
	}
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
		},
		"j": {
			"i": true,
		},
		"k": {
			"j": true,
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
	drop()
	err := buildUpRealCase()
	if err != nil {
		t.Fatal(err)
	}
	queries := []string{
		fmt.Sprintf("INSERT INTO USERS(ID, FIRST_NAME, LAST_NAME, EMAIL, ADDRESS, CREATED_AT) VALUES ('%v', '%v', '%v', '%v', '%v', '%v')", uuid.New(), "some", "name", "test@gmail.com", "SOMETHING", time.Now().Format("2006-01-02 15:04:05")),
		fmt.Sprintf("INSERT INTO USERS(ID, FIRST_NAME, LAST_NAME, EMAIL, ADDRESS, CREATED_AT) VALUES ('%v', '%v', '%v', '%v', '%v', '%v')", uuid.New(), "some", "name", "test2@gmail.com", "SOMETHING", time.Now().Format("2006-01-02 15:04:05")),
	}
	err = database.RunUnsafeQueries(db, queries)
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
	err = database.RunUnsafeQueries(db, queries)
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
