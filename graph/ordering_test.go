package graph_test

import (
	"database/sql"
	"errors"
	"fmt"
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/graph"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/peterldowns/pgtestdb"
	"github.com/peterldowns/pgtestdb/migrators/golangmigrator"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"
)

var db *sql.DB

var pgConf pgtestdb.Config

var migrator pgtestdb.Migrator

var migrationDir = "../db/migrations/"

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

func TestNewOrdering(t *testing.T) {
	migrator = golangmigrator.New(migrationDir + "case1")
	db = pgtestdb.New(t, pgConf, migrator)
	_, err := graph.NewOrdering(nil)
	if err == nil {
		t.Fatal("expected error for nil db")
	}

	_, err = graph.NewOrdering(db)
	if err != nil {
		t.Fatal(err)
	}
}

func TestOrdering_FindOrderCase1(t *testing.T) {
	// case where something on level 2 is moved down to level 4
	migrator = golangmigrator.New(migrationDir + "case1")
	db = pgtestdb.New(t, pgConf, migrator)
	ordering := graph.Ordering{}
	err := ordering.Init(db)
	if err != nil {
		t.Fatal(err)
	}
	order, err := ordering.GetOrder("a")
	if err != nil {
		t.Fatal(err)
	}
	correctOrder := []string{"b", "d", "c", "a"}
	if len(order) != len(correctOrder) {
		t.Fatalf("order.Len() = %d, want %d", len(order), len(correctOrder))
	}

	flag := true

	for i, tableName := range order {
		flag = tableName == correctOrder[i]
		if !flag {
			t.Errorf("Incorect order. Correct order should be %s Instead got %s", strings.Join(correctOrder, ","), strings.Join(order, ","))
		}
	}
}

func TestOrdering_FindOrderCase2(t *testing.T) {
	// case where there's a cyclic dependency
	migrator = golangmigrator.New(migrationDir + "case2")
	db = pgtestdb.New(t, pgConf, migrator)
	ordering := graph.Ordering{}
	err := ordering.Init(db)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ordering.GetOrder("team_members")
	properError := errors.As(err, &graph.CyclicError{})
	if !properError || err == nil {
		t.Errorf("Cyclic error not detected between tables teams and students")
	}
}

func TestOrdering_FindOrderCase3(t *testing.T) {
	// give the function a table with no relationships
	migrator = golangmigrator.New(migrationDir + "case3")
	db = pgtestdb.New(t, pgConf, migrator)

	ordering := graph.Ordering{}
	err := ordering.Init(db)
	if err != nil {
		t.Fatal(err)
	}
	order, err := ordering.GetOrder("users")
	if err != nil {
		t.Errorf("Unexpected Error: %s", err.Error())
	} else if order[0] != "users" {
		t.Errorf("Missing root table")
	}
}

func TestOrdering_FindOrderCase4(t *testing.T) {
	migrator = golangmigrator.New(migrationDir + "case4")
	db = pgtestdb.New(t, pgConf, migrator)
	ordering := graph.Ordering{}
	err := ordering.Init(db)
	if err != nil {
		t.Fatal(err)
	}
	order, err := ordering.GetOrder("team_members")
	if err == nil {
		t.Fatal("Missing table error should have occurred")
	}
	// I only nil check because IDE was being annoying about it
	if order != nil && len(order) != 0 {
		t.Errorf("Empty schema given so the length should be 0 but it isn't. Length is %d", len(order))
	}
}

func TestOrdering_GetCycles(t *testing.T) {
	migrator = golangmigrator.New(migrationDir + "case9")
	db = pgtestdb.New(t, pgConf, migrator)
	ordering, err := graph.NewOrdering(db)
	if err != nil {
		t.Fatal(err)
	}
	expectedLen := 6
	cycles := ordering.GetCycles()
	actualLen := len(cycles)
	if actualLen != expectedLen {
		t.Fatalf("cycles length is %d, expected %d", actualLen, expectedLen)
	}

}

func TestOrdering_GetCyclesForTable(t *testing.T) {
	migrator = golangmigrator.New(migrationDir + "case9")
	db = pgtestdb.New(t, pgConf, migrator)
	ord, err := graph.NewOrdering(db)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ord.GetCyclesForTable("team_members")
	if !errors.As(err, &graph.MissingTableError{}) {
		t.Fatalf("expected 'MissingTableError' got %v", err)
	}

	cycles, err := ord.GetCyclesForTable("b")
	if err != nil {
		t.Fatal(err)
	}
	for _, cycle := range cycles {
		arr := strings.Split(cycle, " --> ")
		if !slices.Contains(arr, "b") {
			t.Fatalf("cycle '%s' does not involve target table 'b'", cycle)
		}
	}
}

func TestOrdering_GetSuggestionQueries(t *testing.T) {
	migrator = golangmigrator.New(migrationDir + "case9")
	db = pgtestdb.New(t, pgConf, migrator)
	ord, err := graph.NewOrdering(db)
	if err != nil {
		t.Fatal(err)
	}
	queries := ord.GetSuggestionQueries()
	err = database.RunUnsafeQueries(db, queries, false)
	if err != nil {
		t.Fatal(err)
	}
	// reset ord
	err = ord.Init(db)
	if err != nil {
		t.Fatal(err)
	}
	cycles := ord.GetCycles()
	if len(cycles) > 0 {
		t.Fatalf("cycles length is %d, expected 0", len(cycles))
	}
}

// check if data is preserved
func TestOrdering_GetSuggestionQueriesDataPreserve(t *testing.T) {
	migrator = golangmigrator.New(migrationDir + "case9")
	db = pgtestdb.New(t, pgConf, migrator)
	// put existing data
	queries := []string{
		"INSERT INTO I(IKEY, JKEY) VALUES (1, NULL)",
		"INSERT INTO J(JKEY, KKEY) VALUES (1, NULL)",
		"INSERT INTO K(KKEY, IKEY) VALUES (1, NULL)",
		"UPDATE I SET JKEY=1 WHERE IKEY=1",
		"UPDATE J SET KKEY=1 WHERE JKEY=1",
		"UPDATE K SET IKEY=1 WHERE KKEY=1",
		"INSERT INTO B(BKEY, BKEY_REF) VALUES (1, NULL)",
		"INSERT INTO B(BKEY, BKEY_REF) VALUES (2, NULL)",
		"UPDATE B SET BKEY_REF=2 WHERE BKEY=1",
	}
	err := database.RunUnsafeQueries(db, queries, false)
	if err != nil {
		t.Fatal(err)
	}
	ord, err := graph.NewOrdering(db)
	if err != nil {
		t.Fatal(err)
	}
	queries = ord.GetSuggestionQueries()
	err = database.RunUnsafeQueries(db, queries, false)
	if err != nil {
		t.Fatal(err)
	}
	// reset ord
	err = ord.Init(db)
	if err != nil {
		t.Fatal(err)
	}
	cycles := ord.GetCycles()
	if len(cycles) > 0 {
		t.Fatalf("cycles length is %d, expected 0", len(cycles))
	}
	// verifications
	allTables := database.GetTableMap(db)
	err = checkCountForTable(db, "b_b", 1)
	if err != nil {
		t.Fatal(err)
	}
	// based on naming convention it's one of these
	allPossibilities := []string{"i_j", "j_k", "k_i", "j_i", "k_j", "i_k"}
	for i, possibility := range allPossibilities {
		_, ok := allTables[possibility]
		if !ok {
			if i == len(allPossibilities)-1 {
				allKeys := slices.Collect(maps.Keys(allTables))
				t.Fatalf("no possibilities were matched in all tables keys [%s] from the array [%s]", strings.Join(allKeys, ", "), strings.Join(allPossibilities, ", "))
			}
		} else {
			err = checkCountForTable(db, possibility, 1)
			if err != nil {
				t.Fatal(err)
			}
			break // break loop to avoid error
		}
	}
}

func TestOrdering_GetSuggestionQueriesForCycles(t *testing.T) {
	migrator = golangmigrator.New(migrationDir + "case9")
	db = pgtestdb.New(t, pgConf, migrator)
	ord, err := graph.NewOrdering(db)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ord.GetCyclesForTable("nonexistant")
	if !errors.As(err, &graph.MissingTableError{}) {
		t.Fatalf("expected 'MissingTableError' got %v", err)
	}

	cycles, err := ord.GetCyclesForTable("b")
	if err != nil {
		t.Fatal(err)
	}
	queries := ord.GetSuggestionQueriesForCycles(cycles)
	err = database.RunUnsafeQueries(db, queries, false)
	if err != nil {
		t.Fatal(err)
	}
	// reset ord
	err = ord.Init(db)
	if err != nil {
		t.Fatal(err)
	}
	cycles, err = ord.GetCyclesForTable("b")
	if err != nil {
		t.Fatal(err)
	}
	if len(cycles) > 0 {
		t.Fatalf("cycles length is %d, expected 0", len(cycles))
	}

	// check to see if it got rid of all cycles (it shouldn't have)
	cycles = ord.GetCycles()
	if len(cycles) != 3 {
		t.Fatalf("cycles length is %d, expected 2", len(cycles))
	}
}

func TestOrdering_GetAndResolveCycles(t *testing.T) {
	// for compound tables
	migrator = golangmigrator.New(migrationDir + "case8")
	db = pgtestdb.New(t, pgConf, migrator)
	ord, err := graph.NewOrdering(db)
	if err != nil {
		t.Fatal(err)
	}
	cycles := ord.GetCycles()
	if len(cycles) == 0 {
		t.Fatal("cycles should exist in schema before being resolved")
	}
	ord.GetAndResolveCycles()
	err = ord.Init(db)
	if err != nil {
		t.Fatal(err)
	}
	cycles = ord.GetCycles()
	if len(cycles) > 0 {
		t.Fatalf("cycles length is %d, expected 0", len(cycles))
	}
}

func TestOrdering_ResolveGivenCycles(t *testing.T) {
	migrator = golangmigrator.New(migrationDir + "case9")
	db = pgtestdb.New(t, pgConf, migrator)
	ord, err := graph.NewOrdering(db)
	if err != nil {
		t.Fatal(err)
	}

	cycles, err := ord.GetCyclesForTable("b")
	if err != nil {
		t.Fatal(err)
	}
	ord.ResolveGivenCycles(cycles)
	// reset ord
	err = ord.Init(db)
	if err != nil {
		t.Fatal(err)
	}
	cycles, err = ord.GetCyclesForTable("b")
	if err != nil {
		t.Fatal(err)
	}
	if len(cycles) > 0 {
		t.Fatalf("cycles length is %d, expected 0", len(cycles))
	}

	// check to see if it got rid of all cycles (it shouldn't have)
	cycles = ord.GetCycles()
	if len(cycles) != 3 {
		t.Fatalf("cycles length is %d, expected 2", len(cycles))
	}
}

func TestOrdering_Len2CycleEdgeCase(t *testing.T) {
	migrator = golangmigrator.New(migrationDir + "case9")
	db = pgtestdb.New(t, pgConf, migrator)
	ord, err := graph.NewOrdering(db)
	if err != nil {
		t.Fatal(err)
	}

	cycles, err := ord.GetCyclesForTable("l")
	if err != nil {
		t.Fatal(err)
	}
	ord.ResolveGivenCycles(cycles)
	lMap := database.GetColumnMap(db, "l")
	mMap := database.GetColumnMap(db, "m")
	if _, ok := lMap["mkey"]; ok {
		t.Fatal("column 'mkey' should have been removed from table 'l'")
	}
	if _, ok := mMap["lkey"]; ok {
		t.Fatal("column 'lkey' should have been removed from table 'm'")
	}
}

// see if our stylistic changes affect data types on new table
func TestStylisticEquality(t *testing.T) {
	var match string
	migrator = golangmigrator.New(migrationDir + "stylistic")
	db = pgtestdb.New(t, pgConf, migrator)
	oldData := database.GetRawColumnMap(db, "a")
	ord, err := graph.NewOrdering(db)
	if err != nil {
		t.Fatal(err)
	}
	ord.GetAndResolveCycles()
	allTables := database.GetAllColumnData(db)
	// based on naming convention it's one of these
	allPossibilities := []string{"a_b", "b_a"}
	for i, possibility := range allPossibilities {
		_, ok := allTables[possibility]
		if !ok {
			if i == len(allPossibilities)-1 {
				allKeys := slices.Collect(maps.Keys(allTables))
				t.Fatalf("no possibilities were matched in all tables keys [%s] from the array [%s]", strings.Join(allKeys, ", "), strings.Join(allPossibilities, ", "))
			}
		} else {
			match = possibility
			break // break loop to avoid error
		}
	}
	newData := database.GetRawColumnMap(db, match)
	for oldName, oldColInfo := range oldData {
		if oldName != "bkey" {
			colInfo, ok := newData[fmt.Sprintf("a_%s", oldName)]
			if !ok {
				t.Fatalf("column 'a_%s' not found in new table", oldName)
			}
			if reflect.DeepEqual(colInfo, oldColInfo) {
				t.Fatalf("columns don't match\nold data = %+v\nnew data = %+v", oldColInfo, colInfo)
			}
		}
	}

}
