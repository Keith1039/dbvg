package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func checkFileContents(t *testing.T, path string, expected []string) {
	path = filepath.Clean(strings.TrimSpace(path)) // WriteQueriesToFile does the same preprocessing, so I'll add some here
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
	err := WriteQueriesToFile(path1, message1)
	if err != nil {
		t.Fatal(err)
	}
	checkFileContents(t, path1, message1) // evaluate the file's contents

	// evaluate second test case
	err = WriteQueriesToFile(path2, message2)
	if err != nil {
		t.Fatal(err)
	}
	checkFileContents(t, path2, message2) // evaluate the file's contents

	// evaluate third test case
	err = WriteQueriesToFile(path3, message3)
	if err != nil {
		t.Fatal(err)
	}
	checkFileContents(t, path3, message3) // evaluate the file's contents

	err = WriteQueriesToFile(path4, []string{})
	if err == nil { // no file should have caused an error
		t.Fatalf("path %s should have caused an error", path4)
	}

	err = WriteQueriesToFile(path5, []string{})
	if err == nil { // empty string should have returned an error
		t.Fatalf("empty string should have caused an error to be returned")
	}
}
