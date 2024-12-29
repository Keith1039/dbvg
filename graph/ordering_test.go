package graph

import (
	"container/list"
	"database/sql"
	"fmt"
	"github.com/Keith1039/Capstone_Test/db"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"log"
	"os"
	"testing"
)

var database *sql.DB

const path = "file://../db/migrations/"

func drop() {
	// drop the database
	driver, err := postgres.WithInstance(database, &postgres.Config{})
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

func init() {
	var err error
	err = os.Setenv("DATABASE_URL", "postgres://postgres:localDB12@localhost:5432/testgres?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	database, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	drop() // drop the database

}

func buildUp(caseName string) error {
	// migrate the schema up
	driver, err := postgres.WithInstance(database, &postgres.Config{})
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

func TestOrdering_FindOrderCase1(t *testing.T) {
	// case where something on level 2 is moved down to level 4
	defer drop()
	caseName := "case1"
	err := buildUp(caseName)
	if err != nil {
		t.Fatal(err)
	}
	relationships := db.CreateRelationshipsWithDB(database)
	ordering := Ordering{AllTables: db.GetTableMap(), AllRelations: relationships, Stack: list.New()}
	order, err := ordering.FindOrder("a")
	if err != nil {
		t.Fatal(err)
	}
	if order["a"] != 1 || order["b"] != 4 || order["c"] != 2 || order["d"] != 3 {
		t.Errorf("Incorect order. Correct order should be a:1, b:4, c:2, d:3. Instead got a:%d, b:%d, c:%d, d:%d.", order["a"], order["b"], order["c"], order["d"])
	}
}

func TestOrdering_FindOrderCase2(t *testing.T) {
	// case where there's a cyclic dependency
	defer drop()
	caseName := "case2"
	err := buildUp(caseName)
	if err != nil {
		t.Fatal(err)
	}
	relationships := db.CreateRelationshipsWithDB(database)
	ordering := Ordering{AllTables: db.GetTableMap(), AllRelations: relationships, Stack: list.New()}
	_, err = ordering.FindOrder("team_members")
	if err == nil {
		t.Errorf("Cyclic error not detected between tables teams and students")
	}
}

func TestOrdering_FindOrderCase3(t *testing.T) {
	// give the function a table with no relationships
	defer drop()
	caseName := "case3"
	err := buildUp(caseName)
	if err != nil {
		t.Fatal(err)
	}
	relationships := db.CreateRelationshipsWithDB(database)
	ordering := Ordering{AllTables: db.GetTableMap(), AllRelations: relationships, Stack: list.New()}
	order, err := ordering.FindOrder("users")
	if err != nil {
		t.Errorf("Unexpected Error: %s", err.Error())
	} else if order["users"] != 1 {
		t.Errorf("Missing root table")
	}
}

func TestOrdering_FindOrderCase4(t *testing.T) {
	defer drop()
	caseName := "case4"
	err := buildUp(caseName)
	if err != nil {
		t.Fatal("Error should have been given")
	}
	relationships := db.CreateRelationshipsWithDB(database)
	ordering := Ordering{AllTables: db.GetTableMap(), AllRelations: relationships, Stack: list.New()}
	order, err := ordering.FindOrder("team_members")
	if err == nil {
		t.Fatal("Missing table error should have occurred")
	}
	if len(order) != 0 {
		fmt.Println(order)
		t.Errorf("Empty schema given so the length should be 0 but it isn't. Length is %d", len(order))
	}
}
