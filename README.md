# DBVG (Work in progress)

Database validator and generator (dbvg) for Postgres. Use as a CLI or import as a library.

__[CLI Documentation](cmd/README.md)__

__[Go Documentation](https://pkg.go.dev/github.com/Keith1039/dbvg)__

## Installation
Currently, there are two ways to get the CLI.

### Using Go Install
`go install github.com/Keith1039/dbvg@latest`

### Downloading via Releases on GitHub
https://github.com/Keith1039/dbvg/releases

**Warning:** If you are on Windows it is not recommended to use this method.
Antivirus' on Windows sometimes flag Golang compiled executables as malware.
This is shown [here](https://go.dev/doc/faq#virus).

## Main Offering
dbvg provides tools to detect/resolve cycles in a database schema
as well as generate a variable amount of table entries while maintaining
table relationships.

This helps the developer by allowing them to avoid making and updating 
scripts that manually create table entries for their database.

This tool is intended for use in a new personal project or for helping to create a 
proof of concept. This tool is designed to be used in a database with 
little to no table entries. This tool was designed with a test environment in mind,
not a production environment.

### Validation
The validation provided by dbvg is cycle aversion and resolution. As databases grow, 
it becomes easy to inadvertently create cyclic relationships between tables. 
This can be averted with proper planning, but in cases where time is limited, such as hackathons
or hack days, this is often skipped. 

This library offers a way to handle this for you. This allows you to work on
the more important aspects of your project while having confidence in your schema.

### Data Generation
As a database grows, it becomes harder to generate test data for it due to the table relationships.

One solution to this problem is to create scripts that generate manual table entries. 
The consequence of this approach is the technical debt of maintaining the script.

Another solution is to use real data for testing. With this, you don't need to worry about
the table relationships, and you have realistic data to use for testing. The consequence of this approach
is that if any changes are made to the schema, it might take time for you to receive new test data. 
Another consequence is that, for you to get real data, you need users for your application. Depending
on the scope of your project, getting users might prove difficult.

With this library, you can allow the code to handle test data generation and focus on more
the finer aspects of your project.

## Basic Usage (As a library)

### Verify if a specific table in the database is part of a cycle [[schema used]](./db/migrations/case9/000001_omni_test_case.up.sql)
```go
package main

import (
	"database/sql"
	"fmt"
	"github.com/Keith1039/dbvg/graph"
	"log"
	"os"
	"strings"
)

func main() {
	var cycles []string
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL")) // open the database connection
	// check error
	if err != nil {
		log.Fatal(err)
	}
	// the name of the table we check for, the function is case-insensitive so "B", "b" " b" etc are the same input
	tableName := "B"
	ord, err := graph.NewOrdering(db)                   // get a new ordering struct
    if err != nil {
        log.Fatal(err)
    }
	cycles, err = ord.GetCyclesForTable(tableName) // get the actual cycles
	if err != nil {
		log.Fatal(err) // print error if it happens
	}
	size := len(cycles) // size of the array
	// format and print the output
	if size > 0 {
		fmt.Printf("The table '%s' is involved in %d cycles: \n%s", tableName, size, strings.Join(cycles, "\n"))
	} else {
		fmt.Printf("The table '%s' is not involved in any cycles.", tableName)
	}
	defer db.Close() // close database connection
}
```
sample output:
```
The table 'B' is involved in 3 cycles: 
b --> b
b --> c --> a --> b
b --> d --> e --> b
```

### Verify if your database schema has cyclical relationships [[schema used]](./db/migrations/case8/000001_create_compound_table.up.sql)

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

	ord, err := graph.NewOrdering(db) // get a new ordering struct
	if err != nil {
	    log.Fatal(err)
	}
	cycles := ord.GetCycles()    // get a linked list of cycles
	// loop through and print cycles
	for _, cycle := range cycles {
		fmt.Println(cycle)
	}

	defer db.Close()  // close database connection
}
```
Sample output:
```
b --> d --> e --> b
a --> b --> c --> a
```

### Remove All cyclical relationships [[schema used]](./db/migrations/case8/000001_create_compound_table.up.sql)
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

	ord, err := graph.NewOrdering(db) // get a new ordering struct
	if err != nil {
	    log.Fatal(err)
	}
	suggestions := ord.GetSuggestionQueries()         // get an array of the queries to be run on the database
	err = database.RunQueriesVerbose(db, suggestions) // runs the queries while printing them
	if err != nil {
		log.Fatal(err) // log the error and close
	}

	defer db.Close() // close the database connection
}
```

sample output:
```
Query 1: CREATE TABLE IF NOT EXISTS b_a(
         b_bkey INT4,
         b_bkey2 INT4,
         a_akey INT4,
        FOREIGN KEY (b_bkey, b_bkey2) REFERENCES b(bkey, bkey2),
        FOREIGN KEY (a_akey) REFERENCES a(akey),
        PRIMARY KEY (b_bkey, b_bkey2, a_akey)
)
Query 2: INSERT INTO b_a(b_bkey, b_bkey2, a_akey)
SELECT b.bkey, b.bkey2, a.akey
FROM a
INNER JOIN b
ON a.bref = b.bkey AND a.bref2 = b.bkey2;
Query 3: ALTER TABLE a DROP COLUMN bref2;
Query 4: ALTER TABLE a DROP COLUMN bref;
Query 5: CREATE TABLE IF NOT EXISTS b_e(
         b_bkey INT4,
         b_bkey2 INT4,
         e_ekey INT4,
        FOREIGN KEY (b_bkey, b_bkey2) REFERENCES b(bkey, bkey2),
        FOREIGN KEY (e_ekey) REFERENCES e(ekey),
        PRIMARY KEY (b_bkey, b_bkey2, e_ekey)
)
Query 6: INSERT INTO b_e(b_bkey, b_bkey2, e_ekey)
SELECT b.bkey, b.bkey2, e.ekey
FROM e
INNER JOIN b
ON e.bref = b.bkey AND e.bref2 = b.bkey2;
Query 7: ALTER TABLE e DROP COLUMN bref;
Query 8: ALTER TABLE e DROP COLUMN bref2;
```

### Generate entries for a table [[schema used]](./db/real_migrations/000001_shop_example.up.sql)
```go
package main

import (
	"database/sql"
	"fmt"
	"github.com/Keith1039/dbvg/parameters"
	_ "github.com/lib/pq"
	"log"
	"os"
)

func main() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL")) // open the database connection
	if err != nil {
		log.Fatal(err)
	}

	writer, err := parameters.NewQueryWriter(db, "purchases")
	if err != nil {
		log.Fatal(err) // log error
	}
	insertBatch, deleteBatch := writer.GenerateEntries(1) // functional equivalent to calling writer.GenerateEntry()

	err = insertBatch.Exec(db, true) // run the insert queries from the batch while printing them out
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(".................................................") // print a divider
	err = deleteBatch.Exec(db, true)                                 // run the delete queries from the batch while printing them out
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close() // close the database connection
}
```
sample output:
```
executing query 1: 'INSERT INTO users (id, first_name, last_name, email, address, created_at) VALUES ($1, $2, $3, $4, $5, $6);' with parameters: ['0760dba1-abfe-48ad-a989-a40e3918c118', 'hygGjbWauP', 'WxQkcYQsup', 'ZFQHFrHhNr', 'gXAXDTihuW', '2026-04-12 16:32:48']
executing query 2: 'INSERT INTO companies (id, name, email, created_at) VALUES ($1, $2, $3, $4);' with parameters: ['69d97c9f-3610-4448-b97a-0cf72c9a06cb', 'aKyDTmDVwg', 'AuvwxnEACm', '2026-04-12 16:32:48']
executing query 3: 'INSERT INTO products (description, created_at, id, company_id, item_name, price, quantity) VALUES ($1, $2, $3, $4, $5, $6, $7);' with parameters: ['zPeNhYFZct', '2026-04-12 16:32:48', 'd9fbd334-7c27-4c36-861e-72165b98892c', '69d97c9f-3610-4448-b97a-0cf72c9a06cb', 'ACXgzI', '9.997602235779429', '1']
executing query 4: 'INSERT INTO purchases (product_id, quantity, created_at, user_id) VALUES ($1, $2, $3, $4);' with parameters: ['d9fbd334-7c27-4c36-861e-72165b98892c', '1', '2026-04-12 16:32:48', '0760dba1-abfe-48ad-a989-a40e3918c118']
.................................................
executing query 1: 'DELETE FROM purchases WHERE product_id=$1 AND quantity=$2 AND created_at=$3 AND user_id=$4;' with parameters: ['d9fbd334-7c27-4c36-861e-72165b98892c', '1', '2026-04-12 16:32:48', '0760dba1-abfe-48ad-a989-a40e3918c118']
executing query 2: 'DELETE FROM products WHERE description=$1 AND created_at=$2 AND id=$3 AND company_id=$4 AND item_name=$5 AND price=$6 AND quantity=$7;' with parameters: ['zPeNhYFZct', '2026-04-12 16:32:48', 'd9fbd334-7c27-4c36-861e-72165b98892c', '69d97c9f-3610-4448-b97a-0cf72c9a06cb', 'ACXgzI', '9.997602235779429', '1']
executing query 3: 'DELETE FROM companies WHERE id=$1 AND name=$2 AND email=$3 AND created_at=$4;' with parameters: ['69d97c9f-3610-4448-b97a-0cf72c9a06cb', 'aKyDTmDVwg', 'AuvwxnEACm', '2026-04-12 16:32:48']
executing query 4: 'DELETE FROM users WHERE id=$1 AND first_name=$2 AND last_name=$3 AND email=$4 AND address=$5 AND created_at=$6;' with parameters: ['0760dba1-abfe-48ad-a989-a40e3918c118', 'hygGjbWauP', 'WxQkcYQsup', 'ZFQHFrHhNr', 'gXAXDTihuW', '2026-04-12 16:32:48']
```
*Note*: The `QueryWriter` struct cannot be used if a cycle exists in the path for the desired table.
It is recommended to always resolve cycles before generating data. below is the result of using the above
code on a schema that has cycles.

[[schema used]](./db/migrations/case8/000001_create_compound_table.up.sql)
```
2025/02/18 15:27:55 error, the following cycles have been detected in the database schema: b --> d --> e --> b | b --> c --> a --> b
exit status 1
```
