package main

import (
	"fmt"
	"github.com/Keith1039/Capstone_Test/db"
	"github.com/Keith1039/Capstone_Test/graph"
	"github.com/Keith1039/Capstone_Test/parameters"
)

func main() {

	writer := parameters.QueryWriter{TableName: "students"}
	err := writer.Init()
	//writer := parameters.QueryWriter{TableName: "students"}
	//err := writer.Init()
	//if err != nil {
	//	panic(err)
	//}
	//writer.ProcessTables()
	//e := writer.InsertQueryQueue.Front()
	//for e != nil {
	//	fmt.Println(e.Value.(string))
	//	e = e.Next()
	//}
	var ord graph.Ordering
	var queryWriter parameters.QueryWriter

	ord.Init()
	queryWriter.TableName = "b"
	queryWriter.Init()

	fmt.Println("Detecting cycles....")
	cycles := ord.GetCycles()
	node := cycles.Front()
	for node != nil {
		fmt.Println(fmt.Sprintf("cycle detected: %s", node.Value.(string)))
		node = node.Next()
	}
	fmt.Println("Generating queries to break cycle while maintaining relationships...")
	problemTables := ord.GetCycleBreakingOrder(cycles)
	queries := queryWriter.CreateSuggestions(problemTables) // create the suggestions
	node = queries.Front()
	i := 1
	// print out the queries
	for node != nil {
		fmt.Println(fmt.Sprintf("query %d: %s", i, node.Value.(string)))
		i++
		node = node.Next()
	}
	fmt.Println("Running queries...")
	err = db.RunQueries(queries) // run the queries
	if err != nil {
		panic(err)
	} else {
		fmt.Println("Finished running queries!")
	}
	fmt.Println("Checking if cycles still exist...")
	ord.Init()               // reset the maps
	cycles = ord.GetCycles() // check for cycles
	if cycles.Len() == 0 {
		fmt.Println("No cycles found!")
	} else {
		fmt.Println("Cycles found! Guess this didn't work...")
	}

}
