package main

import (
	"database/sql"
	"fmt"
	database "github.com/Keith1039/Capstone_Test/db"
	"github.com/Keith1039/Capstone_Test/graph"
	"github.com/Keith1039/Capstone_Test/parameters"
	"log"
	"os"
)

var db *sql.DB

func init() {
	var err error
	err = os.Setenv("DATABASE_URL", "postgres://postgres:localDB12@localhost:5432/testgres?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
}

func main() {
	//test_run()
	test_run()
}

func test_run() {
	ord := graph.NewOrdering(db)
	queryWriter, err := parameters.NewQueryWriterFor(db, "b")
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
	queries := ord.GetSuggestionQueries() // create the suggestions
	node = queries.Front()
	i := 1
	// print out the queries
	for node != nil {
		fmt.Println(fmt.Sprintf("query %d: %s", i, node.Value.(string)))
		i++
		node = node.Next()
	}
	fmt.Println("Running queries...")
	err = database.RunQueries(db, queries) // run the queries
	if err != nil {
		panic(err)
	} else {
		fmt.Println("Finished running queries!")
	}
	fmt.Println("Checking if cycles still exist...")
	ord.Init(db)             // reset the maps
	cycles = ord.GetCycles() // check for cycles
	if cycles.Len() == 0 {
		fmt.Println("No cycles found!")
	} else {
		log.Fatal("Cycles found! Guess this didn't work...")
	}
	queryWriter, err = parameters.NewQueryWriterFor(db, "b")
	if err == nil {
		fmt.Println("QueryWriter can now be generated for table 'b'")
	} else {
		panic(err)
	}
	for z := 0; z < 1; z++ {
		queryWriter.ProcessTables()
	}

	fmt.Println("Running insert queries...")
	node = queryWriter.InsertQueryQueue.Front()
	i = 1
	for node != nil {
		fmt.Println(fmt.Sprintf("insert query %d: %s", i, node.Value.(string)))
		node = node.Next()
		i++
	}

	err = database.RunQueries(db, queryWriter.InsertQueryQueue)
	if err != nil {
		panic(err)
	}
	fmt.Println("Insert queries ran successfully!")
	fmt.Println(".............................................")
	fmt.Println("Printing Deletion Queries.....")
	i = 1
	node = queryWriter.DeleteQueryQueue.Front()
	for node != nil {
		fmt.Println(fmt.Sprintf("delete query %d: %s", i, node.Value.(string)))
		i++
		node = node.Next()
	}
	err = database.RunQueries(db, queryWriter.DeleteQueryQueue)
	if err != nil {
		panic(err)
	}
	fmt.Println("Delete queries ran successfully!")
}
