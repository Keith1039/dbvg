# DBVG (Work in progress)

Database validator and generator (dbvg) for Postgres. Use as a CLI or import as a library.

__[CLI Documentation](cmd/README.md)__

__[Go Documentation]()__

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

	ord := graph.NewOrdering(db) // get a new ordering struct
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

	ord := graph.NewOrdering(db)                      // get a new ordering struct
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

### Generate entries for a table [[schema used]](./db/real_migrations/000001_shop_example.up.sql)
```go
package main

import (
	"database/sql"
	"fmt"
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/parameters"
	_ "github.com/lib/pq"
	"log"
	"os"
)

func main() {
	os.Setenv("DATABASE_URL", "postgres://postgres:localDB12@localhost:5432/testgres?sslmode=disable")
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL")) // open the database connection
	if err != nil {
		log.Fatal(err)
	}

	writer, err := parameters.NewQueryWriter(db, "purchases")
	if err != nil {
		log.Fatal(err) // log error
	}
	insertQueries, deleteQueries := writer.GenerateEntries(1) // functional equivalent to calling writer.GenerateEntry()

	err = database.RunQueriesVerbose(db, insertQueries) // run the insert queries while printing them out
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(".................................................") // print a divider
	err = database.RunQueriesVerbose(db, deleteQueries)              // run the delete queries to delete the inserted values
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close() // close the database connection
}
```
sample output:
```
Query 1: INSERT INTO companies (id, name, email, created) VALUES ('59dff505-fca0-4703-b9dc-28b257b2e83f', 'RfsZsjvnAB', 'FbbKXnnzyu', '2025-03-06 19:36:36');
Query 2: INSERT INTO products (id, company_id, item_name, price, quantity, description, created) VALUES ('165f178b-d246-42c1-9f76-13d4d06132ec', '59dff505-fca0-4703-b9dc-28b257b2e83f', 'MZCEDV', 60.30811117342869::MONEY, 1, 'YorqnDvEHk', '2025-03-06 19:36:36');
Query 3: INSERT INTO users (email, address, created, id, name, last_name) VALUES ('ARRRMBbaUP', 'pmbBxRDhHZ', '2025-03-06 19:36:36', 'c97e0b57-89f1-4605-baa3-3a7922cb4800', 'WzpvjaQtFQ', 'YttUJbPifL');
Query 4: INSERT INTO purchases (user_id, product_id, quantity, created) VALUES ('c97e0b57-89f1-4605-baa3-3a7922cb4800', '165f178b-d246-42c1-9f76-13d4d06132ec', 1, '2025-03-06 19:36:36');
.................................................
Query 1: DELETE FROM purchases WHERE user_id='c97e0b57-89f1-4605-baa3-3a7922cb4800' AND product_id='165f178b-d246-42c1-9f76-13d4d06132ec' AND quantity=1 AND created='2025-03-06 19:36:36';
Query 2: DELETE FROM users WHERE email='ARRRMBbaUP' AND address='pmbBxRDhHZ' AND created='2025-03-06 19:36:36' AND id='c97e0b57-89f1-4605-baa3-3a7922cb4800' AND name='WzpvjaQtFQ' AND last_name='YttUJbPifL';
Query 3: DELETE FROM products WHERE id='165f178b-d246-42c1-9f76-13d4d06132ec' AND company_id='59dff505-fca0-4703-b9dc-28b257b2e83f' AND item_name='MZCEDV' AND price=60.30811117342869::MONEY AND quantity=1 AND description='YorqnDvEHk' AND created='2025-03-06 19:36:36';
Query 4: DELETE FROM companies WHERE id='59dff505-fca0-4703-b9dc-28b257b2e83f' AND name='RfsZsjvnAB' AND email='FbbKXnnzyu' AND created='2025-03-06 19:36:36';
```
*Note*: The `QueryWriter` struct cannot be used if a cycle exists in the path for the desired table.
It is recommended to always resolve cycles before generating data. below is the result of using the above
code on a schema that has cycles.

[[schema used]](./db/migrations/case8/000001_create_compound_table.up.sql)
```
2025/02/18 15:27:55 error, the following cycles have been detected in the database schema: b --> d --> e --> b | b --> c --> a --> b
exit status 1
```
