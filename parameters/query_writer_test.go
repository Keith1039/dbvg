package parameters_test

import (
	"errors"
	"github.com/Keith1039/dbvg/graph"
	"github.com/Keith1039/dbvg/parameters"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/golangmigrator"
	"testing"
)

const migrationDir = "../db/migrations/"

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

func TestNewQueryWriter_Generic(t *testing.T) {
	migrator = golangmigrator.New(migrationDir + "case9")
	db = pgtestdb.New(t, pgConf, migrator)
	_, err := parameters.NewQueryWriter(db, "some_table")
	if !errors.As(err, &graph.MissingTableError{}) {
		t.Fatalf("expected 'MissingTableError', got %v", err)
	}
	_, err = parameters.NewQueryWriter(db, "b")
	if !errors.As(err, &graph.CyclicError{}) {
		t.Fatalf("expected 'CyclicError', got %v", err)
	}
	_, err = parameters.NewQueryWriterWithTemplate(nil, "b", "some_path")
	if err == nil {
		t.Fatal("error should have due to nil DB connection")
	}
	// path error
	_, err = parameters.NewQueryWriterWithTemplate(db, "z", "some_path")
	if err == nil {
		t.Fatal("error should happen due to to non existent file at path ")
	}
}

func TestQueryWriter_GenerateEntries(t *testing.T) {
	migrator = golangmigrator.New(realMigrationDir)
	db = pgtestdb.New(t, pgConf, migrator)
	writer, err := parameters.NewQueryWriter(db, "purchases")
	if err != nil {
		t.Fatal(err)
	}
	expectedAmount := defaultAmount * len(writer.TableOrder)
	insertBatch, deleteBatch := writer.GenerateEntries(defaultAmount)
	if insertBatch.Size() != expectedAmount {
		t.Fatalf("insertBatch.Size() returned %d instead of %d", insertBatch.Size(), expectedAmount)
	}
	if deleteBatch.Size() != expectedAmount {
		t.Fatalf("deleteBatch.Size() returned %d instead of %d", deleteBatch.Size(), expectedAmount)
	}
}

func TestQueryWriter_GenerateEntry(t *testing.T) {
	migrator = golangmigrator.New(realMigrationDir)
	db = pgtestdb.New(t, pgConf, migrator)
	writer, err := parameters.NewQueryWriter(db, "purchases")
	if err != nil {
		t.Fatal(err)
	}
	expectedAmount := len(writer.TableOrder)
	insertBatch, deleteBatch := writer.GenerateEntry()
	if insertBatch.Size() != expectedAmount {
		t.Fatalf("insertBatch.Size() returned %d instead of %d", insertBatch.Size(), expectedAmount)
	}
	if deleteBatch.Size() != expectedAmount {
		t.Fatalf("deleteBatch.Size() returned %d instead of %d", deleteBatch.Size(), expectedAmount)
	}
}

func BenchmarkGenerateQueries(b *testing.B) {
	migrator = golangmigrator.New(realMigrationDir)
	db = pgtestdb.New(b, pgConf, migrator)
	writer, err := parameters.NewQueryWriter(db, "purchases")
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for range b.N {
		insertBatch, deleteBatch := writer.GenerateEntries(500)
		b.Logf("\ninsert batch size: %d\ndelete batch size: %d", insertBatch.Size(), deleteBatch.Size())
	}
	b.StopTimer()
}
