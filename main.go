package main

import (
	"fmt"
	"github.com/Keith1039/Capstone_Test/db"
	"github.com/Keith1039/Capstone_Test/graph"
	"github.com/Keith1039/Capstone_Test/parameters"
	"log"
)

func main() {

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

	ord := graph.NewOrdering()
	queryWriter, err := parameters.NewQueryWriterFor("b")
	if err != nil {
		fmt.Println("Cannot create QueryWriter for table 'b' while cycles exist")
	}
	fmt.Println("Detecting cycles....")
	cycles := ord.GetCycles()
	node := cycles.Front()
	for node != nil {
		fmt.Println(fmt.Sprintf("cycle detected: %s", node.Value.(string)))
		node = node.Next()
	}
	fmt.Println("Generating queries to break cycle while maintaining relationships...")
	problemTables := ord.GetCycleBreakingOrder(cycles)
	queries := ord.CreateSuggestions(problemTables) // create the suggestions
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
		log.Fatal("Cycles found! Guess this didn't work...")
	}
	queryWriter, err = parameters.NewQueryWriterFor("b")
	if err == nil {
		fmt.Println("QueryWriter can now be generated for table 'b'")
	} else {
		panic(err)
	}
	queryWriter.ProcessTables()
	fmt.Println("Running insert queries...")
	node = queryWriter.InsertQueryQueue.Front()
	i = 1
	for node != nil {
		fmt.Println(fmt.Sprintf("insert query %d: %s", i, node.Value.(string)))
		node = node.Next()
		i++
	}

	err = db.RunQueries(queryWriter.InsertQueryQueue)
	if err != nil {
		panic(err)
	}
	fmt.Println("Insert queries ran successfully!")

}
