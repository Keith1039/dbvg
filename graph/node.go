package graph

type topologicalNode struct {
	TableName     string
	path          string
	RelatedTables []string
	slider        int
	visited       bool
	completed     bool
}

func getTopologicalNodes(allTables map[string]int, allRelations map[string]map[string]map[string]string) map[string]*topologicalNode {
	m := make(map[string]*topologicalNode) // map of the table names tied to the node
	for tableName := range allTables {
		relations, exists := allRelations[tableName]
		if exists {
			temp := make(map[string]bool) // make a temporary map
			for _, relation := range relations {
				table := relation["Table"] // take the table
				temp[table] = true         // if it exists, who cares it's a map
			}
			arr := make([]string, len(temp))
			slider := 0
			for key := range temp {
				arr[slider] = key
				slider++
			}
			node := &topologicalNode{TableName: tableName, RelatedTables: arr}
			m[tableName] = node
		} else {
			m[tableName] = &topologicalNode{TableName: tableName, completed: true}
		}
	}
	return m
}
