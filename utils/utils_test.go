package utils_test

import (
	"container/list"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Keith1039/dbvg/graph"
	"github.com/Keith1039/dbvg/utils"
	"github.com/dromara/carbon/v2"
	"github.com/google/uuid"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/golangmigrator"
	regen "github.com/zach-klippenstein/goregen"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

var db *sql.DB

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
func getTempDirAndFile(t *testing.T) (*os.File, string) {
	dir := t.TempDir()
	f, err := os.CreateTemp(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	return f, dir
}

func checkFileContents(t *testing.T, path string, expected []string) {
	path = filepath.Clean(strings.TrimSpace(path)) // utils.WriteQueriesToFile does the same preprocessing, so I'll add some here
	bytes, err := os.ReadFile(path)                // read files
	if err != nil {                                // error check
		t.Fatal(err)
	}
	stringArr := strings.Split(string(bytes), "\n") // get the array
	if !slices.Equal(stringArr, expected) {         // check if the two arrays match
		t.Fatal(fmt.Sprintf("File content %s does not match the expected content of %s", stringArr, expected))
	}
}

func TestWriteQueriesToFile(t *testing.T) {
	dir := t.TempDir()
	path1 := fmt.Sprintf("%s/test1.txt", dir)                       // see if a test file with no message can be handled
	path2 := fmt.Sprintf("   %s/something/some/test2.txt    ", dir) // make multiple sub-subdirectories with whitespace in path
	path3 := fmt.Sprintf("%s/something/test3", dir)                 // see if an extension is enforced
	path4 := fmt.Sprintf("%s/something/test4/", dir)                // this should fail because no file name is specified
	path5 := ""                                                     // see if empty string is handled properly

	message1 := []string{""}
	message2 := []string{"message2", "message2.1"}
	message3 := []string{"message3", "message3.1", "message3.2", "message3.3"}
	// evaluate first test case
	err := utils.WriteQueriesToFile(path1, message1)
	if err != nil {
		t.Fatal(err)
	}
	checkFileContents(t, path1, message1) // evaluate the file's contents

	// evaluate second test case
	err = utils.WriteQueriesToFile(path2, message2)
	if err != nil {
		t.Fatal(err)
	}
	checkFileContents(t, path2, message2) // evaluate the file's contents

	// evaluate third test case
	err = utils.WriteQueriesToFile(path3, message3)
	if err != nil {
		t.Fatal(err)
	}
	checkFileContents(t, path3, message3) // evaluate the file's contents

	err = utils.WriteQueriesToFile(path4, []string{})
	if err == nil { // no file should have caused an error
		t.Fatalf("path %s should have caused an error", path4)
	}

	err = utils.WriteQueriesToFile(path5, []string{})
	if err == nil { // empty string should have returned an error
		t.Fatalf("empty string should have caused an error to be returned")
	}
}

func TestGetTimeFromString(t *testing.T) {
	invalidString := "asd-20021-12"
	nowTime := carbon.Now()
	validString := nowTime.String()
	_, err := utils.GetTimeFromString(invalidString)
	if err == nil {
		t.Fatal("error should have been returned, string was an invalid time")
	}
	_, err = utils.GetTimeFromString(validString)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGenericRetrieveInsertTemplateJSON(t *testing.T) {
	noFilePath := "./db/"
	_, err := utils.RetrieveInsertTemplateJSON(noFilePath)
	if err == nil {
		t.Fatalf("path '%s' should have caused an error since no file was specified", noFilePath)
	}
	fileNoExists := "./db/safmdasofmasf.json"
	_, err = utils.RetrieveInsertTemplateJSON(fileNoExists)
	if err == nil {
		t.Fatalf("path '%s' should have caused an error since file doesn't exist", fileNoExists)
	}
}

func TestWriteInsertTemplateToFileGeneric(t *testing.T) {
	sampleTemplate := map[string]map[string]map[string]any{
		"table": {
			"column": {
				"TYPE": "INT",
				"CoDe": "RANDOM",
			},
		},
	}
	err := utils.WriteInsertTemplateToFile("           ", sampleTemplate)
	if err == nil {
		t.Fatal("error should have been returned since string was empty")
	}

	err = utils.WriteInsertTemplateToFile("../db/migrations/code_test", sampleTemplate)
	if err == nil {
		t.Fatal("error should have been returned since no file was specified")
	}

	err = utils.WriteInsertTemplateToFile("./test_dir/file.json", sampleTemplate)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat("./test_dir/file.json"); os.IsNotExist(err) {
		t.Fatal("file wasn't created in the new dir")
	}
	err = os.RemoveAll("./test_dir")
	if err != nil {
		t.Fatal(err)
	}
}

func TestWriteAndRetrieveInsertTemplateJSON(t *testing.T) {
	f, _ := getTempDirAndFile(t)
	defer f.Close()
	sampleTemplate := map[string]map[string]map[string]any{
		"table": {
			"column": {
				"TYPE": "INT",
				"CoDe": "RANDOM",
			},
		},
	}
	err := utils.WriteInsertTemplateToFile(f.Name(), sampleTemplate)
	if err != nil {
		t.Fatal(err)
	}
	m, err := utils.RetrieveInsertTemplateJSON(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range m {
		for k2, v2 := range v {
			for k3, v3 := range v2 {
				if v3 != sampleTemplate[k][k2][k3] {
					t.Fatalf("error, maps %v  and %v are not equal", sampleTemplate, m)
				}
			}
		}
	}
	jsonData, err := json.MarshalIndent(map[string]any{"something": true}, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(f.Name(), jsonData, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}

	m, err = utils.RetrieveInsertTemplateJSON(f.Name())
	if err == nil {
		t.Fatal("error should have been returned, template was invalid")
	}
}

func TestUpdateInsertTemplateGeneric(t *testing.T) {
	sampleTemplate := map[string]map[string]map[string]any{
		"table": {
			"column": {},
		},
	}
	// generic file doesn't exist
	_, err := utils.UpdateInsertTemplate("./safdasfawfasfasegf.json", sampleTemplate)
	if err == nil {
		t.Fatal(err)
	}
}

func TestUpdateInsertTemplate(t *testing.T) {
	f, _ := getTempDirAndFile(t)
	defer f.Close()
	sampleTemplate := map[string]map[string]map[string]any{
		"table": {
			"column": {},
		},
	}
	err := utils.WriteInsertTemplateToFile(f.Name(), sampleTemplate)
	if err != nil {
		t.Fatal(err)
	}
	_, err = utils.UpdateInsertTemplate(f.Name(), sampleTemplate)
	if err == nil {
		t.Fatal("code key missing, error should have occurred")
	}

	sampleTemplate["table"]["column"]["CoDe"] = "RANDOM"
	err = utils.WriteInsertTemplateToFile(f.Name(), sampleTemplate)
	if err != nil {
		t.Fatal(err)
	}
	_, err = utils.UpdateInsertTemplate(f.Name(), sampleTemplate)
	if err == nil {
		t.Fatal("type key missing, error should have occurred")
	}

	sampleTemplate["table"]["column"]["TYPE"] = "INT"
	err = utils.WriteInsertTemplateToFile(f.Name(), sampleTemplate)
	if err != nil {
		t.Fatal(err)
	}
	_, err = utils.UpdateInsertTemplate(f.Name(), sampleTemplate)
	if err == nil {
		t.Fatal("value key missing, error should have occurred")
	}

	sampleTemplate["table"]["column"]["vaLue"] = any(5)
	_, err = utils.UpdateInsertTemplate(f.Name(), sampleTemplate)
	if err != nil {
		t.Fatal(err)
	}
	sampleClone := maps.Clone(sampleTemplate)
	sampleTemplate["something"] = map[string]map[string]any{
		"column2": {
			"TYPE":  "VARCHAR",
			"CoDe":  "STATIC",
			"value": any("XD"),
		},
	}
	_, err = utils.UpdateInsertTemplate(f.Name(), sampleTemplate)
	if err != nil {
		t.Fatal(err)
	}
	retrievedTemplate, err := utils.RetrieveInsertTemplateJSON(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(retrievedTemplate, sampleTemplate) {
		t.Fatalf("retrieved template '%v'\ninputed template '%v'", retrievedTemplate, sampleTemplate)
	}
	// check to see if irrelevant data is left out
	_, err = utils.UpdateInsertTemplate(f.Name(), sampleClone)
	if err != nil {
		t.Fatal(err)
	}
	retrievedTemplate, err = utils.RetrieveInsertTemplateJSON(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(retrievedTemplate, sampleClone) {
		t.Fatalf("retrieved template '%v'\ninputed template '%v'", retrievedTemplate, sampleClone)
	}
	// check if the new type is saved over the old
	sampleClone["table"]["column"]["TYPE"] = "FLOAT"
	_, err = utils.UpdateInsertTemplate(f.Name(), sampleClone)
	if err != nil {
		t.Fatal(err)
	}
	retrievedTemplate, err = utils.RetrieveInsertTemplateJSON(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(retrievedTemplate, sampleClone) {
		t.Fatalf("retrieved template '%v'\ninputed template '%v'", retrievedTemplate, sampleClone)
	}
}

func TestUpdateInsertTemplateChanges(t *testing.T) {
	f, _ := getTempDirAndFile(t)
	defer f.Close()
	migrator = golangmigrator.New("../db/real_migrations/")
	db = pgtestdb.New(t, pgConf, migrator)
	ord, err := graph.NewOrdering(db)
	if err != nil {
		t.Fatal(err)
	}
	order, err := ord.GetOrder("products")
	if err != nil {
		t.Fatal(err)
	}
	templ := utils.MakeTemplates(db, order)
	// apply changes
	delete(templ, "companies")
	delete(templ["products"], "id")
	templ["products"]["ids"] = map[string]any{
		"TYPE":  "VARCHAR",
		"CoDe":  "STATIC",
		"value": any("XD"),
	}
	templ["deletable"] = map[string]map[string]any{
		"some_col": {
			"TYPE":  "VARCHAR",
			"CoDe":  "STATIC",
			"value": any("XD"),
		},
	}
	err = utils.WriteInsertTemplateToFile(f.Name(), templ)
	if err != nil {
		t.Fatal(err)
	}
	templ = utils.MakeTemplates(db, order)
	changes, err := utils.UpdateInsertTemplate(f.Name(), templ)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes) != 4 {
		t.Fatalf("wrong number of changes, expected 4, got %d", len(changes))
	}
}

func TestListToStringArray(t *testing.T) {
	l := list.New()
	l.PushBack("1")
	l.PushBack("2")
	l.PushFront("3")
	arr := utils.ListToStringArray(l)
	if len(arr) != 3 {
		t.Fatalf("list length should be 3, got %v", len(arr))
	}
	node := l.Front()
	i := 0
	// check the order
	for node != nil {
		nodeVal := node.Value.(string)
		if nodeVal != arr[i] {
			t.Fatalf("arr value should be %v, got %v", nodeVal, arr[i])
		}
		node = node.Next()
		i++
	}
}

func TestMakeTemplates(t *testing.T) {
	migrator = golangmigrator.New("../db/real_migrations/")
	db = pgtestdb.New(t, pgConf, migrator)
	ord, err := graph.NewOrdering(db)
	if err != nil {
		t.Fatal(err)
	}
	order, err := ord.GetOrder("products")
	if err != nil {
		t.Fatal(err)
	}
	expectedTemplate := map[string]map[string]map[string]any{
		"products": {
			"id": {
				"type":  "UUID",
				"code":  "",
				"value": nil,
			},
			"item_name": {
				"type":  "VARCHAR",
				"code":  "",
				"value": nil,
			},
			"price": {
				"type":  "FLOAT",
				"code":  "",
				"value": nil,
			},
			"quantity": {
				"type":  "INT",
				"code":  "",
				"value": nil,
			},
			"description": {
				"type":  "VARCHAR",
				"code":  "",
				"value": nil,
			},
			"created_at": {
				"type":  "DATE",
				"code":  "",
				"value": nil,
			},
		},
		"companies": {
			"id": {
				"type":  "UUID",
				"code":  "",
				"value": nil,
			},
			"name": {
				"type":  "VARCHAR",
				"code":  "",
				"value": nil,
			},
			"email": {
				"type":  "VARCHAR",
				"code":  "",
				"value": nil,
			},
			"created_at": {
				"type":  "DATE",
				"code":  "",
				"value": nil,
			},
		},
	}
	templ := utils.MakeTemplates(db, order)
	for k, v := range expectedTemplate {
		if _, ok := templ[k]; !ok {
			t.Fatalf("missing table '%s' in template", k)
		} else {
			for col, val := range v {
				if _, ok = templ[k][col]; !ok {
					fmt.Println(templ[k])
					t.Fatalf("missing column '%s' for table '%s' in template", col, k)
				} else {
					if !reflect.DeepEqual(val, templ[k][col]) {
						t.Fatalf("expected '%v' for column '%s' of table '%s' but received '%v'", val, col, k, templ[k][col])
					}
				}
			}
		}
	}
}

func TestTrimAndUpperString(t *testing.T) {
	generate, err := regen.Generate("[a-z]{10}")
	if err != nil {
		t.Fatal(err)
	}
	transformString := utils.TrimAndUpperString(generate)
	expectedStr := strings.ToUpper(strings.TrimSpace(generate))
	if transformString != expectedStr {
		t.Fatalf("expected '%s' but got '%s'", expectedStr, transformString)
	}
}

func TestGetStringType(t *testing.T) {
	uu := uuid.New()
	if "uuid.UUID" != utils.GetStringType(uu) {
		t.Fatalf("expected 'uuid.UUID' but got '%s'", utils.GetStringType(uu))
	}
}
