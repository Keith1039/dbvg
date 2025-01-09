package graph

type TopologicalNode struct {
	TableName     string
	path          string
	RelatedTables []string
	slider        int
	visited       bool
	completed     bool
}

type OrderInfoNode struct {
	tableName       string
	parentTableName string
}

type TableNode struct {
	TableName  string
	ColumnData map[string]string
	Parameters map[string]string
}

func GetTopologicalNodes(allTables map[string]int, allRelations map[string]map[string]map[string]string) map[string]*TopologicalNode {
	m := make(map[string]*TopologicalNode) // map of the table names tied to the node
	for tableName := range allTables {
		relations, exists := allRelations[tableName]
		if exists {
			temp := make(map[string]int) // make a temporary map
			for _, relation := range relations {
				table := relation["Table"] // take the table
				temp[table] = 1            // if it exists, who cares it's a map
			}
			arr := make([]string, len(temp))
			slider := 0
			for key, _ := range temp {
				arr[slider] = key
				slider++
			}
			node := &TopologicalNode{TableName: tableName, RelatedTables: arr}
			m[tableName] = node
		} else {
			m[tableName] = &TopologicalNode{TableName: tableName, completed: true}
		}
	}
	return m
}
