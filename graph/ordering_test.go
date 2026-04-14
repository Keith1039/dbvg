package graph_test

import (
	"database/sql"
	"errors"
	"github.com/Keith1039/dbvg/graph"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/golangmigrator"
	"strings"
	"testing"
)

var db *sql.DB

var pgConf pgtestdb.Config

var migrator pgtestdb.Migrator

var migrationDir = "../db/migrations/"

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

func TestOrdering_FindOrderCase1(t *testing.T) {
	// case where something on level 2 is moved down to level 4
	migrator = golangmigrator.New(migrationDir + "case1")
	db = pgtestdb.New(t, pgConf, migrator)
	ordering := graph.Ordering{}
	err := ordering.Init(db)
	if err != nil {
		t.Fatal(err)
	}
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
	migrator = golangmigrator.New(migrationDir + "case2")
	db = pgtestdb.New(t, pgConf, migrator)
	ordering := graph.Ordering{}
	err := ordering.Init(db)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ordering.GetOrder("team_members")
	properError := errors.As(err, &graph.CyclicError{})
	if !properError || err == nil {
		t.Errorf("Cyclic error not detected between tables teams and students")
	}
}

func TestOrdering_FindOrderCase3(t *testing.T) {
	// give the function a table with no relationships
	migrator = golangmigrator.New(migrationDir + "case3")
	db = pgtestdb.New(t, pgConf, migrator)

	ordering := graph.Ordering{}
	err := ordering.Init(db)
	if err != nil {
		t.Fatal(err)
	}
	order, err := ordering.GetOrder("users")
	if err != nil {
		t.Errorf("Unexpected Error: %s", err.Error())
	} else if order[0] != "users" {
		t.Errorf("Missing root table")
	}
}

func TestOrdering_FindOrderCase4(t *testing.T) {
	migrator = golangmigrator.New(migrationDir + "case4")
	db = pgtestdb.New(t, pgConf, migrator)
	ordering := graph.Ordering{}
	err := ordering.Init(db)
	if err != nil {
		t.Fatal(err)
	}
	order, err := ordering.GetOrder("team_members")
	if err == nil {
		t.Fatal("Missing table error should have occurred")
	}
	// I only nil check because IDE was being annoying about it
	if order != nil && len(order) != 0 {
		t.Errorf("Empty schema given so the length should be 0 but it isn't. Length is %d", len(order))
	}
}
