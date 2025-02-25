# DBVG (Work in progress)

Database validator and generator (dbvg) for Postgres. Use as a CLI or import as a library.

__[CLI Documentation](cmd/README.md)__

__[Go Documentation]()__

## Main Offering
This project provides tools to detect/resolve cycles in a database schema.
This project also provides tools to generate X amount of table entries
while maintaining table relationships.

This helps the developer by allowing them to avoid making and updating 
scripts that manually create table entries for their database.

This tool is intended for use in a new personal project or for helping to create a 
proof of concept. This tool is designed to be used in a database with 
little to no table entries. In other words, please don't use this in prod...

### Validation
The validation provided by dbvg is cycle aversion and cycle resolution. As databases grow, 
it becomes easy to inadvertently create cyclic relationships between tables. 
This can be averted with proper planning but in cases where time is limited, such as hackathons
or hack days, this is often skipped. 

This library offers a way to handle this for you. This allows you to work on
the more important aspects of your project while having confidence in your schema.

### Data Generation
As a database grows, it also becomes harder to generate test data for it, due to the table relationships.

One solution to this problem is to create scripts that generate manual table entries. 
The consequence of this approach is the technical debt of maintaining this script.

Another solution is to use real data for testing. With this, you don't need to worry about
the table relationships, and you have realistic data to use for testing. The consequence of this approach
is that if any changes are made to the schema, it might take time for you to receive new test data. 
Another consequence is that, for you to get real data, you need users for your application. Depending
on the scope of your project, getting users for an unfinished product would be difficult.

With this library, you can allow the code to handle test data generation and focus on more
the finer aspects of your project.

## Basic Usage (As a library)

### Verify if your database schema has cyclical relationships

``` go
package main

import (
	"database/sql"
	"fmt"
	"github.com/Keith1039/dbvg/graph"
	"log"
	"os"
)

func main() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL")) // open the database connection
	if err != nil {
		log.Fatal(err)
	}

	ord := graph.NewOrdering(db) // get a new ordering struct
	cycles := ord.GetCycles()    // get a linked list of cycles
	// loop through the list
	node := cycles.Front()
	for node != nil {
		fmt.Println(node.Value.(string)) // print out the cycles
		node = node.Next()
	}

	defer db.Close()
}
```
Sample output:
```
b --> d --> e --> b
a --> b --> c --> a
```

### Remove All cyclical relationships
``` go
package main

import (
	"database/sql"
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/graph"
	"log"
	"os"
)

func main() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL")) // open the database connection
	if err != nil {
		log.Fatal(err)
	}

	ord := graph.NewOrdering(db)                      // get a new ordering struct
	suggestions := ord.GetSuggestionQueries()         // get a linked list of the suggestion queries
	err = database.RunQueriesVerbose(db, suggestions) // runs the suggestion queries and prints them
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}
```

sample output:
```
Query 1: ALTER TABLE e DROP COLUMN bref;
Query 2: ALTER TABLE e DROP COLUMN bref2;
Query 3: CREATE TABLE IF NOT EXISTS b_e(
         b_bkey_ref INT4,
         b_bkey2_ref INT4,
         e_ekey_ref INT4,
        FOREIGN KEY (b_bkey_ref, b_bkey2_ref) REFERENCES b(bkey, bkey2),
        FOREIGN KEY (e_ekey_ref) REFERENCES e(ekey),
        PRIMARY KEY (b_bkey_ref, b_bkey2_ref, e_ekey_ref)
)
Query 4: ALTER TABLE a DROP COLUMN bref;
Query 5: ALTER TABLE a DROP COLUMN bref2;
Query 6: CREATE TABLE IF NOT EXISTS b_a(
         b_bkey_ref INT4,
         b_bkey2_ref INT4,
         a_akey_ref INT4,
        FOREIGN KEY (b_bkey_ref, b_bkey2_ref) REFERENCES b(bkey, bkey2),
        FOREIGN KEY (a_akey_ref) REFERENCES a(akey),
        PRIMARY KEY (b_bkey_ref, b_bkey2_ref, a_akey_ref)
)
```

### Generate X amounts of entries for a table
```go
package main

import (
	"database/sql"
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/parameters"
	"log"
	"os"
)

func main() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL")) // open the database connection
	if err != nil {
		log.Fatal(err)
	}

	writer, err := parameters.NewQueryWriterFor(db, "b")  // create a new query writer for table "b"
	if err != nil {
		log.Fatal(err)
	}
	writer.GenerateEntries(1)  // functional equivalent of writer.GenerateEntry() 

	err = database.RunQueriesVerbose(db, writer.InsertQueryQueue) // run the insert queries
	if err != nil {
		log.Fatal(err)
	}
	err = database.RunQueries(db, writer.DeleteQueryQueue) // run the deletion queries for cleanup (optional)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}
```
sample output:
```
Query 1: INSERT INTO e (ekey) VALUES (0);
Query 2: INSERT INTO d (dkey, eref) VALUES (0, 0);
Query 3: INSERT INTO a (akey) VALUES (0);
Query 4: INSERT INTO c (aref, ckey) VALUES (0, 0);
Query 5: INSERT INTO b (bkey, bkey2, cref, dref) VALUES (0, 0, 0, 0);
Query 1: DELETE FROM b WHERE bkey=0 AND bkey2=0 AND cref=0 AND dref=0;
Query 2: DELETE FROM c WHERE aref=0 AND ckey=0;
Query 3: DELETE FROM a WHERE akey=0;
Query 4: DELETE FROM d WHERE dkey=0 AND eref=0;
Query 5: DELETE FROM e WHERE ekey=0;
```
*Note*: The `QueryWriter` struct cannot be used if a cycle exists in the path for the desired table.
It is recommended to always resolve cycles before generating data. below is the result of using the above
code on a schema that has cycles.
```
2025/02/18 15:27:55 error, the following cycles have been detected in the database schema: b --> d --> e --> b | b --> c --> a --> b
exit status 1
```
