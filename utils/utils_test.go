package utils_test

import (
	"encoding/json"
	"fmt"
	"github.com/Keith1039/dbvg/utils"
	"github.com/golang-module/carbon"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

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

func TestRetrieveInsertTemplateJSON(t *testing.T) {
	dir := t.TempDir()
	f, err := os.CreateTemp(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sampleTemplate := map[string]map[string]map[string]any{
		"table": {
			"column": {
				"TYPE": "INT",
				"CoDe": "RANDOM",
			},
		},
	}
	err = writeMapToJSONFile(f.Name(), sampleTemplate)
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
