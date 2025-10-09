package graph

import (
	"container/list"
	"github.com/Keith1039/dbvg/utils"
	"strings"
)

func cleanCyclicPath(cString string) string {
	allTables := strings.Split(cString, ",")
	cycleTable := allTables[len(allTables)-1] // get the last table
	flag := false
	i := 0
	for i < len(allTables) && !flag {
		flag = allTables[i] == cycleTable // check to see if we found the table
		if !flag {                        // if we didn't find it, increment i
			i++
		}
	}
	// the first string is also the last so we cut the last string off to not have duplicates in string
	return strings.Join(allTables[i:], " --> ")
}
func getTablesMap(cycles []string) map[string]map[string]bool {
	m := make(map[string]map[string]bool)

	for _, cycle := range cycles {
		l := make(map[string]bool)
		cycleArr := strings.Split(cycle, " --> ") // get the array of strings
		for i := 0; i < len(cycleArr)-1; i++ {    // we skip the last since it's a duplicate of the first
			table := cycleArr[i]
			l[table] = true // add the table to the map
		}
		m[cycle] = l // set the array
	}
	return m
}

func getFrequency(cycles []string) map[string]int {
	m := make(map[string]int)
	for _, cycle := range cycles {
		if cycle != "" {
			cycleArr := strings.Split(cycle, " --> ")
			cycleArr = cycleArr[:len(cycleArr)-1]
			for _, table := range cycleArr {
				_, exists := m[table]
				// unnecessary but it makes more sense this way
				if !exists {
					m[table] = 1
				} else {
					m[table] = m[table] + 1
				}
			}
		}
	}
	return m
}

func getMostMentioned(fmap map[string]int) string {
	var k string
	var v int
	// loop over the map
	for key, value := range fmap {
		if value > v { // check if the value of the current key is greater than current
			k = key   // set the new key
			v = value // set the new value
		}
	}
	return k
}

func constructProblemTableMap(problemTable string, tableMap map[string]map[string]bool) map[string]bool {
	relevantTablesMap := make(map[string]bool) // tables entry
	for _, tables := range tableMap {          // go through each map for the cycles
		if _, ok := tables[problemTable]; ok { // if problem table is in the cycle
			for table := range tables { // loop through the keys (table)
				if _, alreadyThere := relevantTablesMap[table]; !alreadyThere { // check if the table already has an entry
					relevantTablesMap[table] = true // add the entry
				}
			}
		}
	}
	return relevantTablesMap // return the problem table map
}

func getRelevantKeys(relations map[string]map[string]map[string]string, tableName string, refTableName string) []string {
	tableRelations := relations[refTableName] // get the map of FKs
	relevantKeys := list.New()                // make a new list

	for col, colRelations := range tableRelations { //loop through the map
		if colRelations["Table"] == tableName && tableName != refTableName { // see if this key is for referencing the same table and ISN'T the same table
			relevantKeys.PushBack(colRelations["Column"]) // add the referenced column to the list
		} else if colRelations["Table"] == tableName { // condition where it is the same table
			relevantKeys.PushBack(col) // add the column to the list
		}
	}
	// by definition this array should be at minimum, size: 1
	return utils.ListToStringArray(relevantKeys) // return an array version of the list
}
