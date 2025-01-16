package graph

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"log"
	"os"
	"strings"
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
	var builder strings.Builder
	defer drop()
	caseName := "case1"
	err := buildUp(caseName)
	if err != nil {
		t.Fatal(err)
	}
	ordering := Ordering{}
	ordering.Init()
	order, err := ordering.FindOrder("a")
	if err != nil {
		t.Fatal(err)
	}
	correctOrder := []string{"b", "d", "c", "a"}
	if order.Len() != len(correctOrder) {
		t.Fatalf("order.Len() = %d, want %d", order.Len(), len(correctOrder))
	}
	i := 0
	flag := true
	node := order.Front()
	for node != nil && flag {
		flag = node.Value.(string) == correctOrder[i]
		i++
		node = node.Next()
	}
	if !flag {
		node = order.Front()
		for node != nil {
			builder.WriteString(node.Value.(string))
			node = node.Next()
		}
		t.Errorf("Incorect order. Correct order should be a:1, b:4, c:2, d:3. Instead got %s", builder.String())
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
	ordering.Init()
	_, err = ordering.FindOrder("team_members")
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
	ordering.Init()
	order, err := ordering.FindOrder("users")
	if err != nil {
		t.Errorf("Unexpected Error: %s", err.Error())
	} else if order.Front().Value.(string) != "users" {
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
	ordering.Init()
	order, err := ordering.FindOrder("team_members")
	if err == nil {
		t.Fatal("Missing table error should have occurred")
	}
	// I only nil check because IDE was being annoying about it
	if order != nil && order.Len() != 0 {
		fmt.Println(order)
		t.Errorf("Empty schema given so the length should be 0 but it isn't. Length is %d", order.Len())
	}
}
