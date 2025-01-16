package main

import (
	"fmt"
	"github.com/Keith1039/Capstone_Test/parameters"
)

func main() {

	writer := parameters.QueryWriter{TableName: "students"}
	err := writer.Init()
	if err != nil {
		panic(err)
	}
	writer.ProcessTables()
	e := writer.InsertQueryQueue.Front()
	for e != nil {
		fmt.Println(e.Value.(string))
		e = e.Next()
	}

}
