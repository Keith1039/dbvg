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
	"maps"
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
	tl.allTables = database.GetTableMap(tl.db)         // get the table map
	tl.allRelations = database.GetRelationships(tl.db) // get the table relations
	tl.stack = list.New()                              // initiate the list
}

// GetCycles uses DFS to detect cycles in the database schema, all detected cycles are added to a linked list and returned as an array
func (tl *Ordering) GetCycles() []string {
	// check if there is a cycle in the entire database schema
	cycles := list.New()                // init the list
	visited := maps.Clone(tl.allTables) // map that keeps track of the tables we've visited (shallow copy of all tables map)
	topologicalNodes := getTopologicalNodes(tl.allTables, tl.allRelations)
	for tableName := range visited { // loop through the map of all tables
		newCycles, localVisited := tl.findCycles(tableName, topologicalNodes) // find the cycles and the visited tables
		cycles.PushBackList(newCycles)                                        // append the new list to the current one
		for local := range localVisited {
			delete(visited, local) // delete visited tables since they have already been covered
		}
	}
	return utils.ListToStringArray(cycles) // return all cycles found
}

// GetCyclesForTable uses DFS to search for any cycles stemming from our root node (i.e. the table we want to verify)
func (tl *Ordering) GetCyclesForTable(tableName string) ([]string, error) {
	var relevantCycles []string
	tableName = utils.TrimAndLowerString(tableName) // make the name case-insensitive
	_, ok := tl.allTables[tableName]
	// check if the table even exists before we run the search
	if !ok {
		return nil, MissingTableError{tableName: tableName} // return missing table error
	}
	// could be a bit faster if I tampered with findCycles so that it only cares if the cycle involves our root node but why bother, it's fast enough as is
	cycles, _ := tl.findCycles(tableName, getTopologicalNodes(tl.allTables, tl.allRelations)) // use dfs starting from the root node (i.e. the table we want to verify)
	// check if we have any cycles
	if cycles.Len() > 0 {
		relevantCycles = getRelevantCycles(tableName, cycles) // trim off any cycles that don't include our target table
	}
	return relevantCycles, nil // return the relevant cycles
}

// GetSuggestionQueries returns a list of queries necessary to remove found cycles in the database schema
func (tl *Ordering) GetSuggestionQueries() []string {
	cycles := tl.GetCycles() // get the cycles
	cycleBreaking, relMap := tl.getCycleBreakingOrder(cycles)
	return tl.getSuggestions(cycleBreaking, relMap) // get and return the suggestions in array format
}

// GetSuggestionQueriesForCycles returns a list of queries necessary to remove a subset of cycles in the database schema
func (tl *Ordering) GetSuggestionQueriesForCycles(cycles []string) []string {
	cycleBreaking, relMap := tl.getCycleBreakingOrder(cycles)
	return tl.getSuggestions(cycleBreaking, relMap) // get and return the suggestions in array format
}

// GetAndResolveCycles immediately runs the suggestion queries instead of returning them unlike GetSuggestionQueries
func (tl *Ordering) GetAndResolveCycles() {
	cycles := tl.GetCycles() // get your cycles
	cycleBreaking, relMap := tl.getCycleBreakingOrder(cycles)
	suggestions := tl.getSuggestions(cycleBreaking, relMap) // get your suggestions
	err := database.RunQueries(tl.db, suggestions)          // run the suggestions
	if err != nil {
		log.Fatal(err) // panic if it fails
	}
}

// ResolveGivenCycles gets suggestion queries for the given cycles and runs them similar to GetAndResolveCycles
func (tl *Ordering) ResolveGivenCycles(cycles []string) {
	cycleBreaking, relMap := tl.getCycleBreakingOrder(cycles)
	suggestions := tl.getSuggestions(cycleBreaking, relMap) // get your suggestions
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

func (tl *Ordering) getCycleBreakingOrder(cycles []string) (*list.List, map[string]map[string]bool) {
	tables := list.New()
	problemTableMap := make(map[string]map[string]bool) // map of all problem tables and the relevant tables
	tablesMap := getTablesMap(cycles)                   // a map that stores which tables are in each cycle
	// copy cycles so that I can still use it
	cyclesCopy := make([]string, len(cycles))
	copy(cyclesCopy, cycles)

	for len(tablesMap) > 0 { // since we delete from tables map whenever a cycle is solved, once it is 0 we can stop looping
		tablesMentioned := getFrequency(cyclesCopy)                                         // we get a map of tables and how often they appear
		mostMentioned := getMostMentioned(tablesMentioned)                                  // get the problem table
		problemTableMap[mostMentioned] = constructProblemTableMap(mostMentioned, tablesMap) // get related tables for given table
		tables.PushBack(mostMentioned)                                                      // add the most mentioned to the list
		// loop through the cycles
		for i, cycle := range cyclesCopy {
			if cycle != "" && tablesMap[cycle][mostMentioned] { // check to see if the cycle is in the most mentioned
				cyclesCopy[i] = ""       // remove the cycle from the array
				delete(tablesMap, cycle) // since the cycle is considered solved, we can remove it from the map
			}
		}
	}
	return tables, problemTableMap
}

func (tl *Ordering) getSuggestions(cycleBreaking *list.List, relMap map[string]map[string]bool) []string {
	var builder strings.Builder
	var dropBuilder strings.Builder
	var foreignKeyBuilder strings.Builder
	var primaryKeyBuilder strings.Builder
	var joinedBuilder strings.Builder
	inverseRelationships := database.GetInverseRelationships(tl.db) // inverse table relationships
	queries := list.New()                                           // make queries list
	node := cycleBreaking.Front()                                   // get the start node
	pkMap := database.GetTablePKMap(tl.db)                          // primary key map
	for node != nil {
		problemTable := node.Value.(string)
		tableRelations := inverseRelationships[problemTable] // map of tables that reference the current table
		colMap := database.GetRawColumnMap(tl.db, problemTable)
		for refTable := range tableRelations {
			if _, actualProblem := relMap[problemTable][refTable]; actualProblem { // check to see if that dependency is part of the cycle
				problemTableKeys := getRelevantKeys(tl.allRelations, problemTable, refTable)
				refTablePKs := pkMap[refTable]
				refColMap := database.GetRawColumnMap(tl.db, refTable)
				newTablePKs := make([]string, len(problemTableKeys)+len(refTablePKs)) // array of the primary keys we'll assign at the end
				newTableSlider := 0
				// first format the string to get rid of the reference column
				appendDropBuilder(&dropBuilder, problemTable, refTable, tl.allRelations)
				if inverseRelationships[refTable][problemTable] && problemTable != refTable { // if the problem table references the ref table and ISN'T the same table
					appendDropBuilder(&dropBuilder, refTable, problemTable, tl.allRelations)
				}
				newTableName := fmt.Sprintf("%s_%s", problemTable, refTable)         // create new relationship table name
				query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(", newTableName) // create the first part of the query
				builder.WriteString(query)
				appendBuilder(&builder, problemTableKeys, problemTable, colMap, newTablePKs, &newTableSlider)                             // append the primary tables columns
				appendBuilder(&builder, refTablePKs, refTable, refColMap, newTablePKs, &newTableSlider)                                   // append the second tables columns
				appendForeignKey(&foreignKeyBuilder, tl.allRelations, problemTable, problemTableKeys, refTable, refTablePKs, newTablePKs) // append the second tables foreign keys
				appendColumnBuilder(&joinedBuilder, tl.allRelations, newTablePKs, refTable, problemTable, problemTableKeys, refTablePKs)
				appendPrimaryKeys(&primaryKeyBuilder, newTablePKs)
				builder.WriteString(foreignKeyBuilder.String())
				builder.WriteString(fmt.Sprintf("\n\tPRIMARY KEY %s\n);", primaryKeyBuilder.String()))
				queries.PushBack(builder.String())
				queries.PushBack(joinedBuilder.String())
				dropQueries := strings.Split(dropBuilder.String(), "\n")
				dropQueries = dropQueries[0 : len(dropQueries)-1] // cut off the end because it's always empty string
				for _, dropQuery := range dropQueries {
					queries.PushBack(dropQuery)
				}

				// reset the string builders
				builder.Reset()
				dropBuilder.Reset()
				foreignKeyBuilder.Reset()
				primaryKeyBuilder.Reset()
				joinedBuilder.Reset()
			}

		}
		node = node.Next() // move to the next node
	}
	return utils.ListToStringArray(queries) // convert to a string array
}
