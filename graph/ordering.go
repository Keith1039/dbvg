package graph

import (
	"container/list"
	"fmt"
	"github.com/Keith1039/Capstone_Test/db"
	"strings"
)

type Ordering struct {
	AllTables    map[string]int
	AllRelations map[string]map[string]map[string]string
	Stack        *list.List
}

func (tl *Ordering) Init() {
	tl.AllTables = db.GetTableMap()
	tl.AllRelations = db.CreateRelationships()
	tl.Stack = list.New()
}

func (tl *Ordering) HasCycles() *list.List {
	// check if there is a cycle in the entire database schema
	cycles := list.New()
	visited := make(map[string]bool)
	for tname := range tl.AllTables {
		visited[tname] = false
	}
	topologicalNodes := GetTopologicalNodes(tl.AllTables, tl.AllRelations)
	for tableName := range visited {
		newCycles, localVisited := tl.findCycles(tableName, topologicalNodes) // find the cycles and the visited tables
		cycles.PushBackList(newCycles)                                        // append the new list to the current one
		for local := range localVisited {
			delete(visited, local) // delete visited tables since they have already been covered
		}
	}
	return cycles // return all cycles found
}

func (tl *Ordering) hasCyclesForTable(tableName string) *list.List {
	// checks if there is a cycle in the path of a given table
	cycles, _ := tl.findCycles(tableName, GetTopologicalNodes(tl.AllTables, tl.AllRelations))
	return cycles
}

func (tl *Ordering) findCycles(tableName string, topologicalNodes map[string]*TopologicalNode) (*list.List, map[string]bool) {
	var nextTable string
	visited := make(map[string]bool) // map of tables we've visited
	node := topologicalNodes[tableName]
	node.path = node.TableName
	cycles := list.New()    // cycles list
	backtrack := list.New() // queue
	backtrack.PushBack(node)
	for backtrack.Len() > 0 {
		node = backtrack.Front().Value.(*TopologicalNode)
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

func (tl *Ordering) Topological(tableName string) (*list.List, error) {
	_, exist := tl.AllTables[tableName]
	if !exist {
		return nil, MissingTableError{tableName}
	}
	topologicalNodes := GetTopologicalNodes(tl.AllTables, tl.AllRelations)
	l := list.New()                                  // create queue
	backtrack := list.New()                          // create queue
	backtrack.PushFront(topologicalNodes[tableName]) // add the first entry
	for backtrack.Len() > 0 {
		node := backtrack.Front().Value.(*TopologicalNode) // get the node pointer
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

func (tl *Ordering) FindOrder(tableName string) (*list.List, error) {
	_, exists := tl.AllTables[tableName] // check if the table exists
	if !exists {
		return nil, MissingTableError{tableName} // return missing table error
	}
	cycles := tl.hasCyclesForTable(tableName) // get all cycles in that tables path
	if cycles.Len() > 0 {
		return nil, CyclicError{cycles: cycles} // return cyclic error
	}
	return tl.Topological(tableName) // return the topological ordering
}

func (tl *Ordering) CycleBreaking(cycles *list.List) *list.List {
	tables := list.New()
	tablesMap := getTablesMap(cycles) // a map that stores which tables are in each cycle
	for cycles.Len() > 0 {
		tablesMentioned := getFrequency(cycles)            // we get a map of tables and how often they appear
		mostMentioned := getMostMentioned(tablesMentioned) // get the problem table
		tables.PushBack(mostMentioned)                     // add the most mentioned to the list
		node := cycles.Front()
		for node != nil {
			nextNode := node.Next() // next node
			// if the most mentioned is in this node, remove the node from the list
			if tablesMap[node.Value.(string)][mostMentioned] {
				cycles.Remove(node)
			}
			node = nextNode
		}
	}

	return tables
}

func (tl *Ordering) CreateSuggestions(cycles *list.List) *list.List {
	var builder strings.Builder
	pkMap := db.GetTablePKMap()
	inverseRelationships := db.CreateInverseRelationships()
	queries := list.New()
	node := cycles.Front()
	for node != nil {
		tableName := node.Value.(string)
		tableRelations := inverseRelationships[tableName]
		colMap := db.GetRawColumnMap(tableName)
		for table, relation := range tableRelations {
			refColMap := db.GetRawColumnMap(table)
			// first format the string to get rid of the reference column
			queries.PushBack(fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;", table, relation["FKColumn"]))
			query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s_%s(\n", tableName, table)
			builder.WriteString(query)
			query = fmt.Sprintf("\t %s %s,\n\t %s %s, ", tableName+"_ref", colMap[relation["Column"]], table+"_ref", refColMap[pkMap[table]])
			builder.WriteString(query)
			query = fmt.Sprintf("\n\tFOREIGN KEY (%s) REFERENCES %s,", tableName+"_ref", tableName)
			builder.WriteString(query)
			query = fmt.Sprintf("\n\tFOREIGN KEY (%s) REFERENCES %s,", table+"_ref", table)
			builder.WriteString(query)
			query = fmt.Sprintf("\n\tPRIMARY KEY (%s, %s)\n)", tableName+"_ref", table+"_ref")
			builder.WriteString(query)
			queries.PushBack(builder.String())
			builder.Reset()

		}
		node = node.Next()
	}
	return queries
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
