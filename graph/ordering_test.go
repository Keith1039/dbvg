package graph

import (
	"database/sql"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"log"
	"os"
	"strings"
	"testing"
)

var db *sql.DB

const path = "file://../db/migrations/"

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
	drop() // drop the database
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

func TestOrdering_FindOrderCase1(t *testing.T) {
	// case where something on level 2 is moved down to level 4
	defer drop()
	caseName := "case1"
	err := buildUp(caseName)
	if err != nil {
		t.Fatal(err)
	}
	ordering := Ordering{}
	ordering.Init(db)
	order, err := ordering.GetOrder("a")
	if err != nil {
		t.Fatal(err)
	}
	correctOrder := []string{"b", "d", "c", "a"}
	if len(order) != len(correctOrder) {
		t.Fatalf("order.Len() = %d, want %d", len(order), len(correctOrder))
	}

	flag := true

	for i, tableName := range order {
		flag = tableName == correctOrder[i]
		if !flag {
			t.Errorf("Incorect order. Correct order should be %s Instead got %s", strings.Join(correctOrder, ","), strings.Join(order, ","))
		}
	}
}

func TestOrdering_FindOrderCase2(t *testing.T) {
	// case where there's a cyclic dependency
	var cyclicError CyclicError
	defer drop()
	caseName := "case2"
	err := buildUp(caseName)
	if err != nil {
		t.Fatal(err)
	}

	ordering := Ordering{}
	ordering.Init(db)
	_, err = ordering.GetOrder("team_members")
	properError := errors.As(err, &cyclicError)
	if !properError || err == nil {
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

	ordering := Ordering{}
	ordering.Init(db)
	order, err := ordering.GetOrder("users")
	if err != nil {
		t.Errorf("Unexpected Error: %s", err.Error())
	} else if order[0] != "users" {
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
	ordering := Ordering{}
	ordering.Init(db)
	order, err := ordering.GetOrder("team_members")
	if err == nil {
		t.Fatal("Missing table error should have occurred")
	}
	// I only nil check because IDE was being annoying about it
	if order != nil && len(order) != 0 {
		t.Errorf("Empty schema given so the length should be 0 but it isn't. Length is %d", len(order))
	}
	err = db.Close()
	if err != nil {
		log.Fatal(err)
	}
}
