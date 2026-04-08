package template_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/graph"
	"github.com/Keith1039/dbvg/strategy"
	"github.com/Keith1039/dbvg/template"
	"github.com/Keith1039/dbvg/utils"
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

var tableData map[string]map[string]string

var requiredTables []string

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
	ord, err := graph.NewOrdering(db)
	if err != nil {
		log.Fatal(err)
	}
	requiredTables, err = ord.GetOrder("template")
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
	_, err = template.NewInsertTemplate(tableData, requiredTables, tempFile.Name()) // see if we can make a template
	if !errors.As(err, &graph.MissingTableError{}) {
		t.Fatalf("expected MissingTableError, received %v", err)
	}

	sampleTemplate["template"] = sampleTemplate["table"] // change name by copying data to new valid key
	delete(sampleTemplate, "table")                      // delete the key

	err = writeMapToJSONFile(tempFile.Name(), sampleTemplate) // overwrite file
	if err != nil {
		t.Fatal(err)
	}

	_, err = template.NewInsertTemplate(tableData, requiredTables, tempFile.Name()) // run TemplateFrom again
	if !errors.As(err, &graph.MissingColumnError{}) {
		t.Fatalf("expected MissingColumnError, received %v", err)
	}

	sampleTemplate["template"]["uuid"] = sampleTemplate["template"]["column"] // change to a valid column
	delete(sampleTemplate["template"], "column")                              // delete the key

	err = writeMapToJSONFile(tempFile.Name(), sampleTemplate) // overwrite file
	if err != nil {
		t.Fatal(err)
	}

	_, err = template.NewInsertTemplate(tableData, requiredTables, tempFile.Name()) // run template from
	if !errors.As(err, &template.SchemaError{}) {
		t.Fatalf("expected schemaError, received %v", err)
	}

	//add the final value to make it a proper schema
	sampleTemplate["template"]["uuid"]["ValuE"] = any(6)
	err = writeMapToJSONFile(tempFile.Name(), sampleTemplate)
	if err != nil {
		t.Fatal(err)
	}

	_, err = template.NewInsertTemplate(tableData, requiredTables, tempFile.Name())
	if !errors.As(err, &strategy.UnexpectedTypeError{}) {
		t.Fatalf("expected unexpectedTypeError, received %v", err)
	}
}

func TestTemplateWithRequiredTables(t *testing.T) {
	sampleTemplate = map[string]map[string]map[string]any{
		"template": {
			"int": {
				"TYPE":  "INT",
				"CoDe":  "RANDOM",
				"Value": []int{5, 20},
			},
		},
		"irrelivant": {
			"key": {
				"TYPE":  "INT",
				"CoDe":  "SERIAL",
				"Value": nil,
			},
		},
	}
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
	err = writeMapToJSONFile(tempFile.Name(), sampleTemplate) // overwrite file
	if err != nil {
		t.Fatal(err)
	}
	// check if missing required table
	_, err = template.NewInsertTemplate(tableData, []string{"template", "some table"}, tempFile.Name())
	if !errors.As(err, &template.MissingRequiredTableError{}) {
		t.Fatalf("expected MissingRequiredTableError, received %v", err)
	}

	tmpl, err := template.NewInsertTemplate(tableData, requiredTables, tempFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	p := tmpl.GetStrategyCodePair("irrelivant", "key")
	if !p.IsEmpty() {
		t.Fatal("table 'irrelivant' is supposed to be ignored as it has no relation to table 'template'")
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
	_, err = template.NewInsertTemplate(tableData, requiredTables, tempFile.Name()) // shouldn't be an error
	if err != nil {
		t.Fatal(err)
	}
	sampleTemplate["template"]["date"]["CoDe"] = "NULL"
	err = writeMapToJSONFile(tempFile.Name(), sampleTemplate)
	if err != nil {
		t.Fatal(err)
	}
	_, err = template.NewInsertTemplate(tableData, requiredTables, tempFile.Name())
	if err != nil {
		t.Fatal(err)
	}
}

func TestOptionalCodes(t *testing.T) {
	var unsupportedErr strategy.UnexpectedTypeError
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
				"CoDe":  "SERIAL",
				"value": nil,
			},
		},
	}
	err = writeMapToJSONFile(tempFile.Name(), sampleTemplate) // write data to file
	if err != nil {
		t.Fatal(err)
	}
	_, err = template.NewInsertTemplate(tableData, requiredTables, tempFile.Name())
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
	_, err = template.NewInsertTemplate(tableData, requiredTables, tempFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// change val to valid case (float 64)
	sampleTemplate["template"]["int"]["value"] = 6.0
	err = writeMapToJSONFile(tempFile.Name(), sampleTemplate) // write data to file
	if err != nil {
		t.Fatal(err)
	}
	_, err = template.NewInsertTemplate(tableData, requiredTables, tempFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// change val to invalid case (string)
	sampleTemplate["template"]["int"]["value"] = ""
	err = writeMapToJSONFile(tempFile.Name(), sampleTemplate) // write data to file
	if err != nil {
		t.Fatal(err)
	}
	_, err = template.NewInsertTemplate(tableData, requiredTables, tempFile.Name())
	if !errors.As(err, &unsupportedErr) {
		t.Fatalf("expected error of type UnexpectedTypeError, received %v", err)
	}
}

func TestInsertTemplate_GetStrategy(t *testing.T) {
	var tmpl *template.InsertTemplate
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
				"CoDe":  "SERIAL",
				"value": nil,
			},
		},
	}
	err = writeMapToJSONFile(tempFile.Name(), sampleTemplate) // write data to file
	if err != nil {
		t.Fatal(err)
	}
	tmpl, err = template.NewInsertTemplate(tableData, requiredTables, tempFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	s := tmpl.GetStrategyCodePair("template", "int")
	if s.IsEmpty() {
		t.Fatal("expected StrategyCodePair to not be empty")
	}

	s = tmpl.GetStrategyCodePair("templAte   ", "int   ")
	if s.IsEmpty() {
		t.Fatal("expected StrategyCodePair to exist due to sanitizing strings")
	}

	s = tmpl.GetStrategyCodePair("xdff", "int")
	if !s.IsEmpty() {
		t.Fatal("expected StrategyCodePair to be empty")
	}

}

func TestInsertTemplateDefaults(t *testing.T) {
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
	sampleTemplate = map[string]map[string]map[string]any{
		"template": {
			"date": {
				"TYPE":  "DATE",
				"CoDe":  "NOW",
				"value": nil,
			},
		},
	}
	supportedTypes := []string{"INT", "FLOAT", "UUID", "DATE", "BOOL", "VARCHAR"}
	for _, supportedType := range supportedTypes {
		sampleTemplate["template"] = map[string]map[string]any{
			utils.TrimAndLowerString(supportedType): {
				"type":  supportedType,
				"code":  "",
				"value": nil,
			},
		}
		err = writeMapToJSONFile(tempFile.Name(), sampleTemplate) // write data to file
		if err != nil {
			t.Fatal(err)
		}
		_, err = template.NewInsertTemplate(tableData, requiredTables, tempFile.Name())
		if err != nil {
			t.Fatal(err)
		}
	}
}
