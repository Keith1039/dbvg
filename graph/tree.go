package graph

import (
	"container/list"
	"fmt"
)

type TableLevels struct {
	AllRelations map[string]map[string]map[string]string
	levelMap     map[string]int // the tables name and the level it's on from the
	Stack        *list.List
}

type CyclicError struct {
	tableName  string
	rTableName string
}

func (e CyclicError) Error() string {
	return fmt.Sprintf("Circular dependency between tables %s and %s detected", e.tableName, e.rTableName)
}

func (tl *TableLevels) FindOrder(tableName string) (map[string]int, error) {
	var err error
	tl.levelMap = make(map[string]int) // refresh the map
	tableNode := TableInfoNode{tableName: tableName, parentTableName: ""}
	tl.Stack.PushBack(tableNode) // push the table node to the back of the stack
	tl.levelMap[tableName] = 1   // shows that it's the root
	for tl.Stack.Len() > 0 {
		err = tl.processNode() // process the nodes in the stack until it's empty (we reached the bottom)
		if err != nil {
			return nil, err
		}
	}
	return tl.levelMap, nil // return the map of the levels
}

// inefficient (nodes that have already been dealt with could get sent back to the queue)
func (tl *TableLevels) processNode() error {
	tableNode := tl.Stack.Front().Value.(TableInfoNode) // get the next table name
	tableName := tableNode.tableName
	level, _ := tl.levelMap[tableName]         // get the level of the entry
	relations, _ := tl.AllRelations[tableName] // find the relations for that table
	for _, details := range relations {
		rTableName := details["Table"]
		rLevel, ok := tl.levelMap[rTableName]
		if !ok { // if there's no entry give it a level entry and add it to the stack
			tl.levelMap[rTableName] = level + 1                                                 // indicate it's further down the tree than it's parent
			tl.Stack.PushBack(TableInfoNode{tableName: rTableName, parentTableName: tableName}) // add it to the stack
		} else {
			// find connection
			if tableNode.parentTableName != rTableName {
				return CyclicError{tableName: tableName, rTableName: rTableName}
			} else if level != rLevel && rLevel < level+1 {
				tl.levelMap[rTableName] = level + 1                                                 // say hey, this should be lowered actually
				tl.Stack.PushBack(TableInfoNode{tableName: rTableName, parentTableName: tableName}) // add it to the stack
			}
		}
	}
	tl.Stack.Remove(tl.Stack.Front()) // remove the front
	//fmt.Println(tl.Stack.Len())
	return nil
}
