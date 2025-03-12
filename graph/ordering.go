// Package graph contains all graphing algorithms used
//
// the graph package contains all functions that relate to graphs such as the DFS implementation as well as the cycle resolution code
package graph

import (
	"container/list"
	"database/sql"
	"fmt"
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/utils"
	"log"
	"strings"
)

// NewOrdering returns the address of a properly initiated Ordering struct
func NewOrdering(db *sql.DB) *Ordering {
	// get a new ordering
	ord := Ordering{}
	ord.Init(db)
	return &ord
}

// Ordering is a struct that contains the information necessary to detect and remove cycles
type Ordering struct {
	db           *sql.DB                                 // the database connection
	allTables    map[string]int                          // a map of all the tables in the database
	allRelations map[string]map[string]map[string]string // all table relationships in a mapped form
	stack        *list.List                              // a stack
}

// Init takes in a database connection and sets all private variables in the Ordering struct
func (tl *Ordering) Init(db *sql.DB) {
	tl.db = db
	tl.allTables = database.GetTableMap(tl.db)
	tl.allRelations = database.GetRelationships(tl.db)
	tl.stack = list.New()
}

// GetCycles uses DFS to detect cycles, all detected cycles are added to a linked list and returned
func (tl *Ordering) GetCycles() []string {
	// check if there is a cycle in the entire database schema
	cycles := list.New()
	visited := make(map[string]bool)
	for tname := range tl.allTables {
		visited[tname] = false
	}
	topologicalNodes := getTopologicalNodes(tl.allTables, tl.allRelations)
	for tableName := range visited {
		newCycles, localVisited := tl.findCycles(tableName, topologicalNodes) // find the cycles and the visited tables
		cycles.PushBackList(newCycles)                                        // append the new list to the current one
		for local := range localVisited {
			delete(visited, local) // delete visited tables since they have already been covered
		}
	}
	return utils.ListToStringArray(cycles) // return all cycles found
}

// GetSuggestionQueries returns a list of queries necessary to remove found cycles in the database schema
func (tl *Ordering) GetSuggestionQueries() []string {
	cycles := tl.GetCycles() // get the cycles
	cycleBreaking := tl.getCycleBreakingOrder(cycles)
	return tl.getSuggestions(cycleBreaking, cycles) // get and return the suggestions in array format
}

// GetAndResolveCycles immediately runs the suggestion queries instead of returning them unlike GetSuggestionQueries
func (tl *Ordering) GetAndResolveCycles() {
	cycles := tl.GetCycles() // get your cycles
	cycleBreaking := tl.getCycleBreakingOrder(cycles)
	suggestions := tl.getSuggestions(cycleBreaking, cycles) // get your suggestions
	err := database.RunQueries(tl.db, suggestions)          // run the suggestions
	if err != nil {
		log.Fatal(err) // panic if it fails
	}
}

func (tl *Ordering) getCyclesForTable(tableName string) *list.List {
	// checks if there is a cycle in the path of a given table
	cycles, _ := tl.findCycles(tableName, getTopologicalNodes(tl.allTables, tl.allRelations))
	return cycles
}

func (tl *Ordering) findCycles(tableName string, topologicalNodes map[string]*topologicalNode) (*list.List, map[string]bool) {
	var nextTable string
	visited := make(map[string]bool) // map of tables we've visited
	node := topologicalNodes[tableName]
	node.path = node.TableName
	cycles := list.New()    // cycles list
	backtrack := list.New() // queue
	backtrack.PushBack(node)
	for backtrack.Len() > 0 {
		node = backtrack.Front().Value.(*topologicalNode)
		node.visited = true // show that we've visited the node
		visited[node.TableName] = true
		if node.completed {
			node.visited = false                // uncheck it to show that we're done with it
			backtrack.Remove(backtrack.Front()) // remove the node
		} else {
			flag := true
			for flag && node.slider < len(node.RelatedTables) {
				nextTable = node.RelatedTables[node.slider] // get the next table name
				node.slider++                               // increment slider
				flag = topologicalNodes[nextTable].visited  // check if the node has been visited in this path
				if flag {
					cyclicPath := node.path + "," + nextTable // make the initial cyclic path
					cyclicPath = cleanCyclicPath(cyclicPath)  // clean the cyclic path
					cycles.PushBack(cyclicPath)               // cycle detected
				} else {
					tempNode := topologicalNodes[nextTable]     // get the next node
					tempNode.path = node.path + "," + nextTable // create the path for the next node
					backtrack.PushFront(tempNode)               // push the next node to the front of the queue
				}
			}
			if node.slider == len(node.RelatedTables) {
				node.completed = true // mark the node as completed
			}
		}
	}
	return cycles, visited
}

func (tl *Ordering) topological(tableName string) (*list.List, error) {
	_, exist := tl.allTables[tableName]
	if !exist {
		return nil, MissingTableError{tableName}
	}
	topologicalNodes := getTopologicalNodes(tl.allTables, tl.allRelations)
	l := list.New()                                  // create queue
	backtrack := list.New()                          // create queue
	backtrack.PushFront(topologicalNodes[tableName]) // add the first entry
	for backtrack.Len() > 0 {
		node := backtrack.Front().Value.(*topologicalNode) // get the node pointer
		if node.completed {                                // check if the node is marked as completed
			l.PushBack(node.TableName)               // add it to the queue we return
			delete(topologicalNodes, node.TableName) // remove the table from the dict
			backtrack.Remove(backtrack.Front())      // remove the current front in queue
		} else {
			newTableNode, exists := topologicalNodes[node.RelatedTables[node.slider]] // get the next related table
			node.slider++                                                             // move slider
			if node.slider == len(node.RelatedTables) {                               // check if we've reached the end
				node.completed = true // set flag to true
			}
			if exists {
				backtrack.PushFront(newTableNode) // add the new node to the queue
			}
		}
	}
	return l, nil
}

// GetOrder returns a list of table names that need entries before the given table can receive an entry alongside any errors that occur
func (tl *Ordering) GetOrder(tableName string) ([]string, error) {
	tableName = utils.TrimAndLowerString(tableName)
	_, exists := tl.allTables[tableName] // check if the table exists
	if !exists {
		return nil, MissingTableError{tableName} // return missing table error
	}
	cycles := tl.getCyclesForTable(tableName) // get all cycles in that tables path
	if cycles.Len() > 0 {
		return nil, CyclicError{cycles: cycles} // return cyclic error
	}
	topologicalOrdering, err := tl.topological(tableName) // return the topological ordering
	if err != nil {
		return nil, err
	}
	return utils.ListToStringArray(topologicalOrdering), nil
}

func cleanCyclicPath(cString string) string {
	allTables := strings.Split(cString, ",")
	cyleTable := allTables[len(allTables)-1] // get the last table
	flag := false
	i := 0
	for i < len(allTables) && !flag {
		flag = allTables[i] == cyleTable // check to see if we found the table
		if !flag {                       // if we didn't find it, increment i
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
func (tl *Ordering) getCycleBreakingOrder(cycles []string) *list.List {
	tables := list.New()

	tablesMap := getTablesMap(cycles) // a map that stores which tables are in each cycle
	remainingCycles := len(cycles)    // the rest of the cycles
	// copy cycles so that I can still use it
	cyclesCopy := make([]string, len(cycles))
	copy(cyclesCopy, cycles)

	for remainingCycles > 0 {
		tablesMentioned := getFrequency(cyclesCopy)        // we get a map of tables and how often they appear
		mostMentioned := getMostMentioned(tablesMentioned) // get the problem table
		tables.PushBack(mostMentioned)                     // add the most mentioned to the list
		// loop through the cycles
		for i, cycle := range cyclesCopy {
			if cycle != "" && tablesMap[cycle][mostMentioned] { // check to see if the cycle is in the most mentioned
				cyclesCopy[i] = "" // remove the cycle from the array
				remainingCycles--  // decrement the value
			}
		}
	}
	return tables
}

func (tl *Ordering) getSuggestions(cycleBreaking *list.List, cycles []string) []string {
	var builder strings.Builder
	var dropBuilder strings.Builder
	var foreignKeyBuilder strings.Builder
	var primaryKeyBuilder strings.Builder
	inverseRelationships := database.GetInverseRelationships(tl.db)
	queries := list.New()
	node := cycleBreaking.Front()
	pkMap := database.GetTablePKMap(tl.db)
	cycleTableMap := getFrequency(cycles)
	for node != nil {
		refTable := node.Value.(string)
		pks := pkMap[refTable]
		tableRelations := inverseRelationships[refTable]
		colMap := database.GetRawColumnMap(tl.db, refTable)
		for problemTable := range tableRelations {
			if _, actualProblem := cycleTableMap[problemTable]; actualProblem { // check to see if that dependency is part of the cycle
				problemTablePks := pkMap[problemTable]
				refColMap := database.GetRawColumnMap(tl.db, problemTable)
				newTablePKs := make([]string, len(pks)+len(problemTablePks)) // array of the primary keys we'll assign at the end
				newTableSlider := 0
				// first format the string to get rid of the reference column
				appendDropBuilder(&dropBuilder, refTable, problemTable, tl.allRelations)
				if inverseRelationships[problemTable][refTable] {
					appendDropBuilder(&dropBuilder, problemTable, refTable, tl.allRelations)
				}
				dropQueries := strings.Split(dropBuilder.String(), "\n")
				dropQueries = dropQueries[0 : len(dropQueries)-1] // cut off the end because it's always empty string
				for _, dropQuery := range dropQueries {
					queries.PushBack(dropQuery)
				}
				query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s_%s(", refTable, problemTable)
				builder.WriteString(query)
				if refTable == problemTable { // self reference in the table
					for _, pk := range pks {
						newPK := fmt.Sprintf("%s_primary_%s", refTable, pk)
						queryString := getColumnQuery(pk, newPK, colMap)
						builder.WriteString(queryString)
						newTablePKs[newTableSlider] = newPK // assign the new pk to the array
						newTableSlider++                    // increment the slider
					}
					appendForeignKey(&foreignKeyBuilder, refTable, newTablePKs[0:len(pks)], pks)
					for _, pk := range pks {
						newPK := fmt.Sprintf("%s_related_%s", refTable, pk)
						queryString := getColumnQuery(pk, newPK, colMap)
						builder.WriteString(queryString)
						newTablePKs[newTableSlider] = newPK // assign the new pk to the array
						newTableSlider++                    // increment slider
					}
					appendForeignKey(&foreignKeyBuilder, refTable, newTablePKs[len(pks):], pks)

				} else {
					appendBuilder(&builder, pks, refTable, colMap, newTablePKs, &newTableSlider)                    // append the primary tables columns
					appendForeignKey(&foreignKeyBuilder, refTable, newTablePKs[:len(pks)], pks)                     // append the foreign keys
					appendBuilder(&builder, problemTablePks, problemTable, refColMap, newTablePKs, &newTableSlider) // append the second tables columns
					appendForeignKey(&foreignKeyBuilder, problemTable, newTablePKs[len(pks):], problemTablePks)     // append the second tables foreign keys
				}
				appendPrimaryKeys(&primaryKeyBuilder, newTablePKs)
				builder.WriteString(foreignKeyBuilder.String())
				builder.WriteString(fmt.Sprintf("\n\tPRIMARY KEY %s\n)", primaryKeyBuilder.String()))
				queries.PushBack(builder.String())
				builder.Reset()
				dropBuilder.Reset()
				foreignKeyBuilder.Reset()
				primaryKeyBuilder.Reset()
			}

		}
		node = node.Next()
	}
	return utils.ListToStringArray(queries) // convert to a string array
}

func getColumnQuery(pk string, newPK string, colMap map[string]*sql.ColumnType) string {
	var queryString string
	precision, scale, ok := colMap[pk].DecimalSize() // check to see if it's a variable float type
	databaseType := colMap[pk].DatabaseTypeName()
	// pretty unnecessary but oh well
	if databaseType == "FLOAT4" || databaseType == "FLOAT8" {
		switch databaseType {
		case "FLOAT4":
			queryString = fmt.Sprintf("\n\t %s %s,", newPK, "REAL")
		case "FLOAT8":
			queryString = fmt.Sprintf("\n\t %s %s,", newPK, "DOUBLE PRECISION")
		}
	} else {
		if ok {
			// NUMERIC types default is (65535, 65531) so we avoid that
			if precision != 65535 && scale != 65531 {
				queryString = fmt.Sprintf("\n\t %s %s(%d, %d),", newPK, databaseType, precision, scale) // add the precision to the new query
			} else {
				queryString = fmt.Sprintf("\n\t %s %s,", newPK, databaseType) // solely for NUMERIC
			}
		} else {
			length, isVarchar := colMap[pk].Length() // get the size of the column
			if isVarchar && databaseType != "TEXT" { // text is excluded from this
				if databaseType == "VARCHAR" && length != -5 { // exclude default varchar
					queryString = fmt.Sprintf("\n\t %s %s(%d),", newPK, databaseType, length)
				} else if databaseType == "BPCHAR" {
					if length != -5 && length != 1 { // exclude default BPCHAR and CHAR
						queryString = fmt.Sprintf("\n\t %s %s(%d),", newPK, databaseType, length)
					} else if length == 1 {
						queryString = fmt.Sprintf("\n\t %s %s,", newPK, "CHAR")
					} else {
						queryString = fmt.Sprintf("\n\t %s %s,", newPK, databaseType)
					}
				}
			} else {
				queryString = fmt.Sprintf("\n\t %s %s,", newPK, databaseType) // in case all other conditions fail
			}
		}
	}
	return queryString
}
func appendBuilder(builder *strings.Builder, pks []string, table string, colMap map[string]*sql.ColumnType, newTablePKs []string, slider *int) {
	// appends the new column name and the datatype
	for _, pk := range pks {
		newPK := fmt.Sprintf("%s_%s_ref", table, pk)
		queryString := getColumnQuery(pk, newPK, colMap)
		builder.WriteString(queryString)
		newTablePKs[*slider] = newPK // assign the new pk to the array
		*slider++                    // increment slider
	}
}

func appendDropBuilder(builder *strings.Builder, refTable string, problemTable string, allRelations map[string]map[string]map[string]string) {
	relations := allRelations[problemTable]
	for column, relation := range relations {
		if relation["Table"] == refTable { // check to see if the fk matches
			builder.WriteString(fmt.Sprintf(fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;\n", problemTable, column)))
		}
	}
}
func appendForeignKey(foreignKeyBuilder *strings.Builder, table string, pks []string, refTablePks []string) {
	// builds a foreign key constraint
	var keyBuilder strings.Builder
	var referenceBuilder strings.Builder

	for i, pk := range pks {
		if keyBuilder.String() == "" {
			keyBuilder.WriteString(pk)
			referenceBuilder.WriteString(refTablePks[i])
		} else {
			keyBuilder.WriteString(fmt.Sprintf(", %s", pk))
			referenceBuilder.WriteString(fmt.Sprintf(", %s", refTablePks[i]))
		}
	}
	foreignKeyBuilder.WriteString(fmt.Sprintf("\n\tFOREIGN KEY (%s) REFERENCES %s(%s),", keyBuilder.String(), table, referenceBuilder.String()))
}

func appendPrimaryKeys(primaryKeyBuilder *strings.Builder, pks []string) {
	// builds the primary key constraint
	primaryKeyBuilder.WriteString("(")
	for _, key := range pks {
		if primaryKeyBuilder.String() == "(" {
			primaryKeyBuilder.WriteString(key)
		} else {
			primaryKeyBuilder.WriteString(fmt.Sprintf(", %s", key))
		}
	}
	primaryKeyBuilder.WriteString(")")
}
