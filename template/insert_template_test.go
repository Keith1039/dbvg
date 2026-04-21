package template_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/Keith1039/dbvg/graph"
	"github.com/Keith1039/dbvg/strategy"
	"github.com/Keith1039/dbvg/template"
	"github.com/Keith1039/dbvg/utils"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/golangmigrator"
	"os"
	"path/filepath"
	"testing"
)

var db *sql.DB

var testCase map[string]map[string]map[string]any

var table string

const path = "../db/migrations/templates"

var pgConf pgtestdb.Config

var migrator pgtestdb.Migrator

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

func initForTest(t *testing.T) {
	migrator = golangmigrator.New(path)
	db = pgtestdb.New(t, pgConf, migrator)
	table = "irrelevant"
	ord, err := graph.NewOrdering(db)
	if err != nil {
		t.Fatal(err)
	}
	tables, err := ord.GetOrder(table)
	if err != nil {
		t.Fatal(err)
	}
	testCase = utils.MakeTemplates(db, tables)
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

func getTempDirAndFile(t *testing.T) (*os.File, string) {
	dir := t.TempDir()
	f, err := os.CreateTemp(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	return f, dir
}

func getTemplate(db *sql.DB, table string) (map[string]map[string]map[string]any, error) {
	table = utils.TrimAndLowerString(table)
	ord, err := graph.NewOrdering(db)
	if err != nil {
		return nil, err
	}
	tables, err := ord.GetOrder(table)
	if err != nil {
		return nil, err
	}
	templ := utils.MakeTemplates(db, tables)
	return templ, nil
}

func writeInvalidTemplateToFile(path string, data map[string]bool) error {
	// by default this will overwrite existing files
	cleanPath, err := utils.CleanFilePath(path) // make sure the path is clean
	if err != nil {
		return err
	}
	dir, fileName := filepath.Split(cleanPath) // split the dir path and the file name
	if dir != "" {                             // check if the dir path is empty string
		if _, err = os.Stat(dir); os.IsNotExist(err) { // check if directory exists
			err = os.MkdirAll(dir, os.ModePerm) // make all directories and subdirectories
			if err != nil {
				return err // log error and exit
			}
		}
	}
	jsonData, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(dir, fileName), jsonData, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// this code is to check if certain cases produce the correct errors (missing table etc)
// this also indirectly tests if the keys are case-insensitive
func TestNewInsertGenericErrors(t *testing.T) {
	var err error
	initForTest(t)
	file, _ := getTempDirAndFile(t)
	defer file.Close()
	testCase, err = getTemplate(db, "irrelevant")
	if err != nil {
		t.Fatal(err)
	}
	// path testing
	testPath := "./"
	// test no file specified
	_, err = template.NewInsertTemplate(db, "template", "./")
	if err == nil {
		t.Fatalf("expected error since path '%s' doesn't specify file", testPath)
	}

	testPath = "some_file.md"
	// test file doesn't exist
	_, err = template.NewInsertTemplate(db, "template", testPath)
	if err == nil {
		t.Fatalf("expected error since file '%s' doesn't exist", testPath)
	}

	testPath = ""
	// test empty string
	_, err = template.NewInsertTemplate(db, "template", testPath)
	if !errors.As(err, &template.MissingPathError{}) {
		t.Fatal("expected error since this function does not allow empty path")
	}

	// valid path but invalid template (i.e. not map[string]map[string]map[string]any)
	err = writeInvalidTemplateToFile(file.Name(), map[string]bool{"": true})
	if err != nil {
		t.Fatal(err)
	}

	_, err = template.NewInsertTemplate(db, "template", testPath)
	if err == nil {
		t.Fatalf("expected error since JSON template at '%s' is invalid", testPath)
	}

	// put a valid template in
	err = utils.WriteInsertTemplateToFile(file.Name(), testCase)
	if err != nil {
		t.Fatal(err)
	}

	// test table doesn't exist
	tableName := "asfasfasfa"
	_, err = template.NewInsertTemplate(db, tableName, file.Name())
	if !errors.As(err, &graph.MissingTableError{}) {
		t.Fatalf("expected error since the table '%s' does not exist", tableName)
	}

	tableName = "irrelevant"
	// test missing table in template
	testCase, err = getTemplate(db, "irrelevant")
	if err != nil {
		t.Fatal(err)
	}
	delete(testCase, "template") // get rid of required table
	err = utils.WriteInsertTemplateToFile(file.Name(), testCase)
	if err != nil {
		t.Fatal(err)
	}
	_, err = template.NewInsertTemplate(db, tableName, file.Name())
	if !errors.As(err, &template.MissingRequiredTableError{}) {
		t.Fatalf("expected error since the template is missing required table '%s' but received '%v'", "template", err)
	}

	// test schema errors

	// add an irrelevant key to schema to see if sign triggers
	testCase, err = getTemplate(db, "irrelevant")
	if err != nil {
		t.Fatal(err)
	}
	testCase["template"]["date"]["new stuff"] = "henlo"
	err = utils.WriteInsertTemplateToFile(file.Name(), testCase)
	if err != nil {
		t.Fatal(err)
	}
	_, err = template.NewInsertTemplate(db, tableName, file.Name())
	if !errors.As(err, &template.SchemaError{}) {
		t.Fatalf("expected error since there is an additional key '%s' in schema but received '%v'", "new stuff", err)
	}

	// add invalid type to schema
	delete(testCase["template"]["date"], "new stuff")
	testCase["template"]["date"]["code"] = 5
	err = utils.WriteInsertTemplateToFile(file.Name(), testCase)
	if err != nil {
		t.Fatal(err)
	}
	_, err = template.NewInsertTemplate(db, tableName, file.Name())
	if !errors.As(err, &template.SchemaError{}) {
		t.Fatal("expected error since 'code' key is expected to be string but is 'int'")
	}

	// test invalid type
	testCase, err = getTemplate(db, "irrelevant")
	if err != nil {
		t.Fatal(err)
	}
	testCase["template"]["date"]["type"] = "INT"
	err = utils.WriteInsertTemplateToFile(file.Name(), testCase)
	if err != nil {
		t.Fatal(err)
	}
	_, err = template.NewInsertTemplate(db, tableName, file.Name())
	if !errors.As(err, &strategy.UnexpectedTypeError{}) {
		t.Fatal("expected error since 'type' key's value is supposed to be 'DATE' but was 'INT'")
	}

	// missing required column
	testCase, err = getTemplate(db, "irrelevant")
	if err != nil {
		t.Fatal(err)
	}
	delete(testCase["template"], "date")
	err = utils.WriteInsertTemplateToFile(file.Name(), testCase)
	if err != nil {
		t.Fatal(err)
	}
	_, err = template.NewInsertTemplate(db, tableName, file.Name())
	if !errors.As(err, &template.MissingRequiredColumnError{}) {
		t.Fatalf("expected error since the template at '%s' is missing required column '%s' for table '%s'", file.Name(), "date", "template")
	}
}

func TestInsertTemplateCycles(t *testing.T) {
	var err error
	initForTest(t)
	file, _ := getTempDirAndFile(t)
	defer file.Close()
	cycleTable := "cycletable"
	testCase, err = getTemplate(db, "template")
	if err != nil {
		t.Fatal(err)
	}
	err = utils.WriteInsertTemplateToFile(file.Name(), testCase)
	if err != nil {
		t.Fatal(err)
	}
	// validation for required tables is done AFTER we get the path so it doesn't
	// matter that the template doesn't match the table
	_, err = template.NewInsertTemplate(db, cycleTable, file.Name())
	if err == nil {
		t.Fatal("expected error since 'cycletable' has a cycle in it's path")
	}
	_, err = template.NewDefaultInsertTemplate(db, cycleTable)
	if err == nil {
		t.Fatal("expected error since 'cycletable' has a cycle in it's path")
	}

}

func TestTemplateWithRequiredTables(t *testing.T) {
	var err error
	initForTest(t)
	file, _ := getTempDirAndFile(t)
	defer file.Close()
	tableName := "IRRELEVANT"
	testCase, err = getTemplate(db, tableName)
	if err != nil {
		t.Fatal(err)
	}
	// test missing table
	delete(testCase, "template")
	err = utils.WriteInsertTemplateToFile(file.Name(), testCase)
	if err != nil {
		t.Fatal(err)
	}
	_, err = template.NewInsertTemplate(db, tableName, file.Name())
	if !errors.As(err, &template.MissingRequiredTableError{}) {
		t.Fatalf("expected error since the required table '%s' is missing but received: '%v'", "template", err)
	}

	testCase, err = getTemplate(db, tableName)
	if err != nil {
		t.Fatal(err)
	}
	delete(testCase["template"], "date")
	err = utils.WriteInsertTemplateToFile(file.Name(), testCase)
	if err != nil {
		t.Fatal(err)
	}
	_, err = template.NewInsertTemplate(db, tableName, file.Name())
	if !errors.As(err, &template.MissingRequiredColumnError{}) {
		t.Fatalf("expected error since the table '%s' is missing required column '%s' but received: '%v'", "template", "date", err)
	}
}

func TestNewInsertTemplate_Codes(t *testing.T) {
	var err error
	initForTest(t)
	file, _ := getTempDirAndFile(t)
	defer file.Close()
	tableName := "IRRELEVANT"
	testCase, err = getTemplate(db, tableName)
	if err != nil {
		t.Fatal(err)
	}
	// test invalid code
	testCase["template"]["date"]["code"] = "NONEXISTENT"
	testCase["template"]["date"]["value"] = nil
	err = utils.WriteInsertTemplateToFile(file.Name(), testCase)
	if err != nil {
		t.Fatal(err)
	}
	_, err = template.NewInsertTemplate(db, tableName, file.Name())
	if !errors.As(err, &strategy.UnsupportedCodeError{}) {
		t.Fatalf("code 'NONEXISTENT' is not defined for type 'DATE' but received '%v'", err)
	}

	// test valid code with invalid value
	testCase["template"]["date"]["code"] = "RANDOM"
	err = utils.WriteInsertTemplateToFile(file.Name(), testCase)
	if err != nil {
		t.Fatal(err)
	}
	_, err = template.NewInsertTemplate(db, tableName, file.Name())
	if err == nil {
		t.Fatal("value 'nil' is not supported for code 'RANDOM', error should have been returned")
	}

	// test valid code with valid value
	testCase, err = getTemplate(db, tableName)
	if err != nil {
		t.Fatal(err)
	}
	testCase["template"]["int"]["code"] = "SERIAL"
	err = utils.WriteInsertTemplateToFile(file.Name(), testCase)
	if err != nil {
		t.Fatal(err)
	}
	_, err = template.NewInsertTemplate(db, tableName, file.Name())
	if err != nil {
		t.Fatal("value 'nil' is supported for code 'SERIAL', error shouldn't have been returned")
	}

}

func TestInsertTemplate_GetStrategy(t *testing.T) {
	var err error
	initForTest(t)
	file, _ := getTempDirAndFile(t)
	defer file.Close()
	tableName := "IRRELEVANT"
	testCase, err = getTemplate(db, tableName)
	if err != nil {
		t.Fatal(err)
	}
	err = writeMapToJSONFile(file.Name(), testCase) // write data to file
	if err != nil {
		t.Fatal(err)
	}
	tmpl, err := template.NewInsertTemplate(db, table, file.Name())
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
