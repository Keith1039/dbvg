package graph

import (
	"container/list"
	"database/sql"
	"fmt"
	database "github.com/Keith1039/Capstone_Test/db"
	"strings"
)

func NewOrdering(db *sql.DB) *Ordering {
	// get a new ordering
	ord := Ordering{}
	ord.Init(db)
	return &ord
}

type Ordering struct {
	db           *sql.DB
	allTables    map[string]int
	allRelations map[string]map[string]map[string]string
	stack        *list.List
}

func (tl *Ordering) Init(db *sql.DB) {
	tl.db = db
	tl.allTables = database.GetTableMap(tl.db)
	tl.allRelations = database.CreateRelationships(tl.db)
	tl.stack = list.New()
}

func (tl *Ordering) GetCycles() *list.List {
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
	return cycles // return all cycles found
}

func (tl *Ordering) GetSuggestionQueries() *list.List {
	cycles := tl.GetCycles() // get the cycles
	cycleBreaking := tl.getCycleBreakingOrder(cycles)
	suggestionQueries := tl.getSuggestions(cycleBreaking) // get the suggestions
	return suggestionQueries                              // return the suggestions
}

func (tl *Ordering) GetAndResolveCycles() {
	cycles := tl.GetCycles() // get your cycles
	cycleBreaking := tl.getCycleBreakingOrder(cycles)
	suggestions := tl.getSuggestions(cycleBreaking) // get your suggestions
	err := database.RunQueries(tl.db, suggestions)  // run the suggestions
	if err != nil {
		panic(err) // panic if it fails
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

func (tl *Ordering) GetOrder(tableName string) (*list.List, error) {
	_, exists := tl.allTables[tableName] // check if the table exists
	if !exists {
		return nil, MissingTableError{tableName} // return missing table error
	}
	cycles := tl.getCyclesForTable(tableName) // get all cycles in that tables path
	if cycles.Len() > 0 {
		return nil, CyclicError{cycles: cycles} // return cyclic error
	}
	return tl.topological(tableName) // return the topological ordering
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
func getTablesMap(cycles *list.List) map[string]map[string]bool {
	m := make(map[string]map[string]bool)
	node := cycles.Front()
	for node != nil {
		l := make(map[string]bool)
		cycleArr := strings.Split(node.Value.(string), " --> ") // get the array of strings
		for i := 0; i < len(cycleArr)-1; i++ {                  // we skip the last since it's a duplicate of the first
			table := cycleArr[i]
			l[table] = true // add the table to the map
		}
		m[node.Value.(string)] = l // set the array
		node = node.Next()         // move to the next
	}
	return m
}

func getFrequency(cycles *list.List) map[string]int {
	m := make(map[string]int)
	node := cycles.Front()
	for node != nil {
		cycleArr := strings.Split(node.Value.(string), " --> ") // get the array of strings
		cycleArr = cycleArr[0 : len(cycleArr)-1]                // cut off the end
		for _, table := range cycleArr {
			_, exists := m[table]
			// unnecessary but it makes more sense this way
			if !exists {
				m[table] = 1
			} else {
				m[table] = m[table] + 1
			}
		}
		node = node.Next() // move to the next node
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
func (tl *Ordering) getCycleBreakingOrder(problemTables *list.List) *list.List {
	tables := list.New()
	tablesMap := getTablesMap(problemTables) // a map that stores which tables are in each cycle
	for problemTables.Len() > 0 {
		tablesMentioned := getFrequency(problemTables)     // we get a map of tables and how often they appear
		mostMentioned := getMostMentioned(tablesMentioned) // get the problem table
		tables.PushBack(mostMentioned)                     // add the most mentioned to the list
		node := problemTables.Front()
		for node != nil {
			nextNode := node.Next()
			// if the most mentioned is in this node, remove the node from the list
			if tablesMap[node.Value.(string)][mostMentioned] {
				problemTables.Remove(node)
			}
			node = nextNode // move to the next node
		}
	}
	return tables
}

func (tl *Ordering) getSuggestions(cycles *list.List) *list.List {
	var builder strings.Builder
	inverseRelationships := database.CreateInverseRelationships(tl.db)
	queries := list.New()
	node := cycles.Front()
	pkMap := database.GetTablePKMap(tl.db)
	for node != nil {
		refTable := node.Value.(string)
		tableRelations := inverseRelationships[refTable]
		colMap := database.GetRawColumnMap(tl.db, refTable)
		for problemTable, relation := range tableRelations {
			refColMap := database.GetRawColumnMap(tl.db, problemTable)
			// first format the string to get rid of the reference column
			queries.PushBack(fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;", problemTable, relation["FKColumn"]))
			query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s_%s(\n", refTable, problemTable)
			builder.WriteString(query)
			query = fmt.Sprintf("\t %s %s,\n\t %s %s, ", refTable+"_ref", colMap[relation["Column"]], problemTable+"_ref", refColMap[pkMap[problemTable]])
			builder.WriteString(query)
			query = fmt.Sprintf("\n\tFOREIGN KEY (%s) REFERENCES %s,", refTable+"_ref", refTable)
			builder.WriteString(query)
			query = fmt.Sprintf("\n\tFOREIGN KEY (%s) REFERENCES %s,", problemTable+"_ref", problemTable)
			builder.WriteString(query)
			query = fmt.Sprintf("\n\tPRIMARY KEY (%s, %s)\n)", refTable+"_ref", problemTable+"_ref")
			builder.WriteString(query)
			queries.PushBack(builder.String())
			builder.Reset()
		}
		node = node.Next()
	}
	return queries
}
