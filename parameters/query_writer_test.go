package parameters_test

import (
	"github.com/Keith1039/dbvg/parameters"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"testing"
)

func buildUpCase(caseName string) error {
	// migrate the schema up
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	m, err2 := migrate.NewWithDatabaseInstance(
		"file://../db/migrations/"+caseName,
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

func TestNewQueryWriter_Generic(t *testing.T) {
	drop()
	err := buildUpCase("case9") // case with cycle
	if err != nil {
		t.Fatal(err)
	}
	_, err = parameters.NewQueryWriter(db, "some_table")
	if err == nil {
		t.Fatal("table doesn't exist in schema, error should have occured")
	}
	_, err = parameters.NewQueryWriter(db, "b")
	if err == nil {
		t.Fatal("error should have occurred due to cycle in schema")
	}
}

func TestQueryWriter_GenerateEntries(t *testing.T) {
	drop()
	err := buildUp()
	if err != nil {
		t.Fatal(err)
	}
	amount := 500
	writer, err := parameters.NewQueryWriter(db, "purchases")
	if err != nil {
		t.Fatal(err)
	}
	expectedAmount := amount * len(writer.TableOrder)
	insertBatch, deleteBatch := writer.GenerateEntries(amount)
	if insertBatch.Size() != expectedAmount {
		t.Fatalf("insertBatch.Size() returned %d instead of %d", insertBatch.Size(), expectedAmount)
	}
	if deleteBatch.Size() != expectedAmount {
		t.Fatalf("deleteBatch.Size() returned %d instead of %d", deleteBatch.Size(), expectedAmount)
	}
}
