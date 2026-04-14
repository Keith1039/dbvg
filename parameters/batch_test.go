package parameters_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Keith1039/dbvg/graph"
	"github.com/Keith1039/dbvg/parameters"
	"github.com/Keith1039/dbvg/utils"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/golangmigrator"
	"log"
	"os"
	"testing"
	"time"
)

var db *sql.DB

var pgConf pgtestdb.Config

var migrator pgtestdb.Migrator

const realMigrationDir = "../db/real_migrations/"

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
	migrator = golangmigrator.New(realMigrationDir)
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

// to get the desired behavior, we need to take the sample template, shove it into a temporary file
func writeMapToJSONFile(filePath string, data map[string]map[string]map[string]any) error {
	jsonData, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}
	err = os.WriteFile(filePath, jsonData, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func TestQueryBatch_Exec(t *testing.T) {
	migrator = golangmigrator.New(realMigrationDir)
	db = pgtestdb.New(t, pgConf, migrator)
	_, err := db.Exec("INSERT INTO COMPANIES(ID, NAME, EMAIL, CREATED_AT) VALUES($1, $2, $3, $4)", uuid.New(), "test", "some@email.com", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	batchWriter, err := parameters.NewQueryWriter(db, "purchases")
	if err != nil {
		t.Fatal(err)
	}
	insertBatch, deleteBatch := batchWriter.GenerateEntries(500)
	err = insertBatch.Exec(db, false)
	if err != nil {
		t.Fatal(err)
	}
	err = checkCountForTable(db, "COMPANIES", 501)
	if err != nil {
		t.Fatal(err)
	}
	err = deleteBatch.Exec(db, false)
	if err != nil {
		t.Fatal(err)
	}
	err = checkCountForTable(db, "COMPANIES", 1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestQueryBatch_ExecRollback(t *testing.T) {
	migrator = golangmigrator.New(realMigrationDir)
	db = pgtestdb.New(t, pgConf, migrator)
	dir := t.TempDir()
	f, err := os.CreateTemp(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	ord, err := graph.NewOrdering(db)
	if err != nil {
		t.Fatal(err)
	}
	order, err := ord.GetOrder("purchases")
	if err != nil {
		t.Fatal(err)
	}
	template := utils.MakeTemplates(db, order)
	template["users"]["email"]["code"] = "STATIC"
	template["users"]["email"]["value"] = "GonnaFail"
	err = writeMapToJSONFile(f.Name(), template)
	if err != nil {
		t.Fatal(err)
	}
	batchWriter, err := parameters.NewQueryWriter(db, "purchases")
	if err != nil {
		t.Fatal(err)
	}
	batchWriter, err = parameters.NewQueryWriterWithTemplate(db, "purchases", f.Name())
	if err != nil {
		t.Fatal(err)
	}
	insertBatch, _ := batchWriter.GenerateEntries(500)
	err = insertBatch.Exec(db, false)
	if err == nil {
		t.Fatal("unique constraint should have been violated due to static code used for email field")
	}
	err = checkCountForTable(db, "USERS", 0)
	if err != nil {
		t.Fatal(err)
	}
	// reset the writer
	batchWriter, err = parameters.NewQueryWriter(db, "purchases")
	if err != nil {
		log.Fatal(err)
	}
}

func TestQueryBatch_ExecContext(t *testing.T) {
	migrator = golangmigrator.New(realMigrationDir)
	db = pgtestdb.New(t, pgConf, migrator)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	batchWriter, err := parameters.NewQueryWriter(db, "purchases")
	if err != nil {
		t.Fatal(err)
	}
	insertBatch, _ := batchWriter.GenerateEntries(5000)
	err = insertBatch.ExecContext(ctx, db, false)
	if err == nil {
		t.Fatal("should have timed out...")
	}
	err = checkCountForTable(db, "COMPANIES", 0)
	if err != nil {
		t.Fatal(err)
	}
}

// tests the migration with every possible supported type and code with their defaults
func TestQueryBatch_OmniCodeTestDefault(t *testing.T) {
	migrator = golangmigrator.New(migrationDir + "code_test")
	db = pgtestdb.New(t, pgConf, migrator)
	batchWriter, err := parameters.NewQueryWriter(db, "test")
	if err != nil {
		t.Fatal(err)
	}
	insertBatch, deleteBatch := batchWriter.GenerateEntries(500)
	err = insertBatch.Exec(db, false)
	if err != nil {
		t.Fatal(err)
	}
	err = deleteBatch.Exec(db, false)
	if err != nil {
		t.Fatal(err)
	}
}

// tests the migration with every possible supported type and code with their template values
func TestQueryBatch_OmniCodeTestWithTemplate(t *testing.T) {
	migrator = golangmigrator.New(migrationDir + "code_test")
	db = pgtestdb.New(t, pgConf, migrator)
	batchWriter, err := parameters.NewQueryWriterWithTemplate(db, "test", migrationDir+"code_test/omni_test.json")
	if err != nil {
		t.Fatal(err)
	}
	insertBatch, deleteBatch := batchWriter.GenerateEntries(500)
	err = insertBatch.Exec(db, false)
	if err != nil {
		t.Fatal(err)
	}
	err = deleteBatch.Exec(db, false)
	if err != nil {
		t.Fatal(err)
	}
}

func Benchmark_ExecInsert(b *testing.B) {
	migrator = golangmigrator.New(realMigrationDir)
	db = pgtestdb.New(b, pgConf, migrator)
	batchWriter, err := parameters.NewQueryWriter(db, "purchases")
	if err != nil {
		b.Fatal(err)
	}
	insertBatch, deleteBatch := batchWriter.GenerateEntries(5000)
	b.ResetTimer()
	for range b.N {
		b.StartTimer()
		err = insertBatch.Exec(db, false)
		if err != nil {
			b.Fatal(err)
		}
		b.StopTimer()
		err = deleteBatch.Exec(db, false)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
}

func Benchmark_ExecDelete(b *testing.B) {
	migrator = golangmigrator.New(realMigrationDir)
	db = pgtestdb.New(b, pgConf, migrator)
	batchWriter, err := parameters.NewQueryWriter(db, "purchases")
	if err != nil {
		b.Fatal(err)
	}
	insertBatch, deleteBatch := batchWriter.GenerateEntries(5000)
	b.ResetTimer()
	for range b.N {
		b.StopTimer()
		err = insertBatch.Exec(db, false)
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()
		err = deleteBatch.Exec(db, false)
		if err != nil {
			b.Fatal(err)
		}
	}
	b.StopTimer()
}
