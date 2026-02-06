package template

import (
	"database/sql"
	"encoding/json"
	"errors"
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/graph"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"log"
	"os"
	"testing"
)

var db *sql.DB

var sampleTemplate map[string]map[string]map[string]any

var insertTemplate = &InsertTemplate{}

var tableData map[string]map[string]string

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
	err = buildUp("templates")
	if err != nil {
		log.Fatal(err)
	}
	tableData = database.GetAllColumnData(db) // set table data
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

// this code is to check if certain cases produce the correct errors (missing table etc)
// this also indirectly tests if the keys are case-insensitive
func TestGenericErrors(t *testing.T) {
	var missingTableError graph.MissingTableError
	var missingColumnError graph.MissingColumnError
	var schemaerr schemaError
	var unexpectedTypeError UnexpectedTypeError
	// check for missing table error
	sampleTemplate = map[string]map[string]map[string]any{
		"table": {
			"column": {
				"TYPE": "INT",
				"CoDe": "RANDOM",
			},
		},
	}
	tempDir := t.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "") // create a temporary file
	defer func(tempFile *os.File) {
		err = tempFile.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(tempFile)
	if err != nil {
		t.Fatal(err)
	}
	err = writeMapToJSONFile(tempFile.Name(), sampleTemplate) // write to temporary file
	if err != nil {
		t.Fatal(err)
	}
	err = insertTemplate.TemplateFrom(tableData, tempFile.Name()) // see if we can make a template
	if !errors.As(err, &missingTableError) {
		t.Fatalf("expected MissingTableError, received %v", err)
	}

	sampleTemplate["template"] = sampleTemplate["table"] // change name by copying data to new valid key
	delete(sampleTemplate, "table")                      // delete the key

	err = writeMapToJSONFile(tempFile.Name(), sampleTemplate) // overwrite file
	if err != nil {
		t.Fatal(err)
	}

	err = insertTemplate.TemplateFrom(tableData, tempFile.Name()) // run TemplateFrom again
	if !errors.As(err, &missingColumnError) {
		t.Fatalf("expected MissingColumnError, received %v", err)
	}

	sampleTemplate["template"]["uuid"] = sampleTemplate["template"]["column"] // change to a valid column
	delete(sampleTemplate["template"], "column")                              // delete the key

	err = writeMapToJSONFile(tempFile.Name(), sampleTemplate) // overwrite file
	if err != nil {
		t.Fatal(err)
	}

	err = insertTemplate.TemplateFrom(tableData, tempFile.Name()) // run template from
	if !errors.As(err, &schemaerr) {
		t.Fatalf("expected schemaError, received %v", err)
	}

	//add the final value to make it a proper schema
	sampleTemplate["template"]["uuid"]["ValuE"] = any(6)
	err = writeMapToJSONFile(tempFile.Name(), sampleTemplate)
	if err != nil {
		t.Fatal(err)
	}

	err = insertTemplate.TemplateFrom(tableData, tempFile.Name())
	if !errors.As(err, &unexpectedTypeError) {
		t.Fatalf("expected unexpectedTypeError, received %v", err)
	}

}

// all override codes ignore values, this test confirms that behavior
// this also applies to the NULL code which is a special code that behaves like an override for any column type
// the usual checks should still run
func TestOverrideCode(t *testing.T) {
	// sample template with an override code
	tempDir := t.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "") // create a temporary file
	defer func(tempFile *os.File) {
		err = tempFile.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(tempFile)
	if err != nil {
		t.Fatal(err)
	}
	sampleTemplate = map[string]map[string]map[string]any{
		"template": {
			"date": {
				"TYPE":  "DATE",
				"CoDe":  "NOW",
				"value": nil,
			},
		},
	}
	err = writeMapToJSONFile(tempFile.Name(), sampleTemplate) // write data to file
	if err != nil {
		t.Fatal(err)
	}
	err = insertTemplate.TemplateFrom(tableData, tempFile.Name()) // shouldn't be an error
	if err != nil {
		t.Fatal(err)
	}
	sampleTemplate["template"]["date"]["CoDe"] = "NULL"
	err = writeMapToJSONFile(tempFile.Name(), sampleTemplate)
	if err != nil {
		t.Fatal(err)
	}
	err = insertTemplate.TemplateFrom(tableData, tempFile.Name())
	if err != nil {
		t.Fatal(err)
	}
}

func TestOptionalCodes(t *testing.T) {
	var unsupportedErr unsupportedTypeError
	tempDir := t.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "")
	if err != nil {
		t.Fatal(err)
	}
	defer func(tempFile *os.File) {
		err = tempFile.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(tempFile)

	// valid case because default applies (nil value implies use default)
	sampleTemplate = map[string]map[string]map[string]any{
		"template": {
			"int": {
				"TYPE":  "INT",
				"CoDe":  "SEQ",
				"value": nil,
			},
		},
	}
	err = writeMapToJSONFile(tempFile.Name(), sampleTemplate) // write data to file
	if err != nil {
		t.Fatal(err)
	}
	err = insertTemplate.TemplateFrom(tableData, tempFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// kinda redundant but in the future updates to preprocessing will allow both cases
	// change val to valid case (int)
	sampleTemplate["template"]["int"]["value"] = 6
	err = writeMapToJSONFile(tempFile.Name(), sampleTemplate) // write data to file
	if err != nil {
		t.Fatal(err)
	}
	err = insertTemplate.TemplateFrom(tableData, tempFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// change val to valid case (float 64)
	sampleTemplate["template"]["int"]["value"] = 6.0
	err = writeMapToJSONFile(tempFile.Name(), sampleTemplate) // write data to file
	if err != nil {
		t.Fatal(err)
	}
	err = insertTemplate.TemplateFrom(tableData, tempFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// change val to invalid case (string)
	sampleTemplate["template"]["int"]["value"] = ""
	err = writeMapToJSONFile(tempFile.Name(), sampleTemplate) // write data to file
	if err != nil {
		t.Fatal(err)
	}
	err = insertTemplate.TemplateFrom(tableData, tempFile.Name())
	if !errors.As(err, &unsupportedErr) {
		t.Fatalf("expected error of type unsupportedTypeError, received %v", err)
	}
}

func TestRequiredCodes(t *testing.T) {
	tempDir := t.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "")
	if err != nil {
		t.Fatal(err)
	}
	defer func(tempFile *os.File) {
		err = tempFile.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(tempFile)
}
