package parameters_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Keith1039/dbvg/graph"
	"github.com/Keith1039/dbvg/parameters"
	"github.com/Keith1039/dbvg/utils"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"log"
	"os"
	"testing"
	"time"
)

var db *sql.DB

const path = "file://../db/real_migrations/"

var batchWriter *parameters.QueryWriter

func drop() {
	// drop the database
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}
	m, err2 := migrate.NewWithDatabaseInstance(
		path,
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
	drop()          // drop the database
	err = buildUp() // build up
	if err != nil {
		log.Fatal(err)
	}
	batchWriter, err = parameters.NewQueryWriter(db, "purchases")
	if err != nil {
		log.Fatal(err)
	}
}

func buildUp() error {
	// migrate the schema up
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	m, err2 := migrate.NewWithDatabaseInstance(
		path,
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
	drop()
	err := buildUp() // build up
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec("INSERT INTO COMPANIES(ID, NAME, EMAIL, CREATED_AT) VALUES($1, $2, $3, $4)", uuid.New(), "test", "some@email.com", time.Now())
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
	drop()
	err := buildUp() // build up
	if err != nil {
		log.Fatal(err)
	}
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
	drop()
	err := buildUp()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
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

func Benchmark_ExecInsert(b *testing.B) {
	drop()
	err := buildUp()
	if err != nil {
		log.Fatal(err)
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
	drop()
	err := buildUp()
	if err != nil {
		log.Fatal(err)
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
