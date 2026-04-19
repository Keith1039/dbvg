*Note*: In all the examples, for the connection string I use an environmental variable, `$URL`. That is
because connections strings tend to be very long, and it would make the commands look less concise.
You can represent the connection string any way you like so long as it is a valid sql connection string.

# Common Uses

## Validate your database Schema using CLI [[schema used]](../db/migrations/case6/)

``` shell
dbvg validate schema --database "$URL" -v
```

Sample output:
```
Cycle Detected!: c --> a --> b --> c
Cycle Detected!: b --> d --> e --> b
2 cycles detected
```

## Resolve cycles using CLI [[schema used]](../db/migrations/case8/000001_create_compound_table.up.sql)
``` shell
dbvg validate schema --database "$URL" --run
```

Sample output:
```
Cycle Detected!: b --> d --> e --> b
Cycle Detected!: a --> b --> c --> a
2 cycles detected

The commands that will run will affect your database schema (creating/dropping tables and moving data)
Would you like to proceed? [Y]es or [N]o: y
Query 1: CREATE TABLE IF NOT EXISTS b_e(
         b_bkey2 INT4,
         b_bkey INT4,
         e_ekey INT4,
        FOREIGN KEY (b_bkey2, b_bkey) REFERENCES b(bkey2, bkey),
        FOREIGN KEY (e_ekey) REFERENCES e(ekey),
        PRIMARY KEY (b_bkey2, b_bkey, e_ekey)
);
Query 2: INSERT INTO b_e(b_bkey2, b_bkey, e_ekey)
SELECT b.bkey2, b.bkey, e.ekey
FROM e
INNER JOIN b
ON e.bref = b.bkey AND e.bref2 = b.bkey2;
Query 3: ALTER TABLE e DROP COLUMN bref;
Query 4: ALTER TABLE e DROP COLUMN bref2;
Query 5: CREATE TABLE IF NOT EXISTS b_a(
         b_bkey2 INT4,
         b_bkey INT4,
         a_akey INT4,
        FOREIGN KEY (b_bkey2, b_bkey) REFERENCES b(bkey2, bkey),
        FOREIGN KEY (a_akey) REFERENCES a(akey),
        PRIMARY KEY (b_bkey2, b_bkey, a_akey)
);
Query 6: INSERT INTO b_a(b_bkey2, b_bkey, a_akey)
SELECT b.bkey2, b.bkey, a.akey
FROM a
INNER JOIN b
ON a.bref = b.bkey AND a.bref2 = b.bkey2;
Query 7: ALTER TABLE a DROP COLUMN bref;
Query 8: ALTER TABLE a DROP COLUMN bref2;
Queries ran successfully
```

## Generate a template [[schema used]](../db/real_migrations/000001_shop_example.up.sql)
``` shell
dbvg format insert-template --create --database "$URL" --path "some_dir/purchase_template.json" --table "purchases"
```

Sample output:
```
template successfully created at 'some_dir/purchase_template.json'
```

## Update an existing template [[schema used]](../db/real_migrations/000001_shop_example.up.sql)
``` shell
dbvg format insert-template --update --database "$URL" --path "some_dir/purchase_template.json" --table "purchases"
```

Sample output:
```
the following changes were applied to the template at path 'some_dir/purchase_template.json':
+ new table 'companies' added
+ new column 'last_name' added to table 'users'
- column 'married' removed from table 'users'
```

For more regarding templates, please see [the documentation for the format sub palette](format/README.md)

## Verify an existing template [[schema used]](../db/real_migrations/000001_shop_example.up.sql)
``` shell
dbvg format insert-template --verify --database "$URL" --path "some_dir/purchase_template.json" --table "purchases"
```

Sample output:
```
2026/04/19 17:57:23 template at 'some_dir/purchase_template.json' failed with error [for column 'last_name' in table 'users': [code 'INVALID_CODE' is not supported for the type 'VARCHAR']]
```

## Generate a table entry [[schema used]](../db/real_migrations/000001_shop_example.up.sql)
``` shell
dbvg insert entry --database "$URL" --amount 1 --default --table "purchases" -v
```

Sample output:
```
Beginning query generation...
Finished generating queries!
Beginning INSERT query execution...
executing query 1: 'INSERT INTO companies (email, created_at, id, name) VALUES ($1, $2, $3, $4);' with parameters: ['bbOaCBuqkS', '2026-04-19 17:53:32.8720535 -0400 EDT m=+0.205130801', 'f8dccfbb-0cd8-4c94-a1ef-6bf27d629b1e', 'cbYsghhwIp']
executing query 2: 'INSERT INTO products (quantity, description, created_at, id, company_id, item_name, price) VALUES ($1, $2, $3, $4, $5, $6, $7);' with parameters: ['1', 'nGmfBGizmP', '2026-04-19 17:53:32.8720535 -0400 EDT m=+0.205130801', 'd61c5d45-a8be-49a0-acb7-22efb92b93ab', 'f8dccfbb-0cd8-4c94-a1ef-6bf27d629b1e', 'JzNNYQ', '6.384035284624385']
executing query 3: 'INSERT INTO users (id, first_name, last_name, email, address, created_at) VALUES ($1, $2, $3, $4, $5, $6);' with parameters: ['9728eefb-47af-4402-85f5-4e29ec132581', 'sCyqGbGqqZ', 'gXcTfBQiJY', 'tvGqMDEEUf', 'pSYXAOisxf', '2026-04-19 17:53:32.8720535 -0400 EDT m=+0.205130801']
executing query 4: 'INSERT INTO purchases (product_id, quantity, created_at, user_id) VALUES ($1, $2, $3, $4);' with parameters: ['d61c5d45-a8be-49a0-acb7-22efb92b93ab', '1', '2026-04-19 17:53:32.8720535 -0400 EDT m=+0.205130801', '9728eefb-47af-4402-85f5-4e29ec132581']
Finished INSERT query execution!
```

### Generate INSERT and DELETE queries (i.e. clean-up flag is true)
``` shell
dbvg insert entry --database "$URL" --amount 1 --default --table "purchases" --clean-up
```

Sample output:
```
Beginning query generation...
Finished generating queries!
Beginning INSERT query execution...
executing query 1: 'INSERT INTO users (id, first_name, last_name, email, address, created_at) VALUES ($1, $2, $3, $4, $5, $6);' with parameters: ['f9074cc1-fa47-429d-a66b-8a82e6c14b45', 'BjIcVGfSKv', 'SzsPEPXOMh', 'gerGVBYVAb', 'exzCOeCLqg', '2026-04-19 17:59:37.2401304 -0400 EDT m=+0.214377701']
executing query 2: 'INSERT INTO companies (created_at, id, name, email) VALUES ($1, $2, $3, $4);' with parameters: ['2026-04-19 17:59:37.2401304 -0400 EDT m=+0.214377701', '45fdcb8a-dba6-44e6-b5f2-70811ba7ebf6', 'iLnnsgWNsD', 'vWFzvKGBmz']
executing query 3: 'INSERT INTO products (id, company_id, item_name, price, quantity, description, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7);' with parameters: ['20ff98cb-9b1f-4568-bd55-4190d99eb6b6', '45fdcb8a-dba6-44e6-b5f2-70811ba7ebf6', 'yiuFpH', '5.038534828802739', '1', 'QsTnSCDzln', '2026-04-19 17:59:37.2401304 -0400 EDT m=+0.214377701']
executing query 4: 'INSERT INTO purchases (user_id, product_id, quantity, created_at) VALUES ($1, $2, $3, $4);' with parameters: ['f9074cc1-fa47-429d-a66b-8a82e6c14b45', '20ff98cb-9b1f-4568-bd55-4190d99eb6b6', '1', '2026-04-19 17:59:37.2401304 -0400 EDT m=+0.214377701']
Finished INSERT query execution!
Press Enter to begin clean up:
Beginning DELETE query execution...
executing query 1: 'DELETE FROM purchases WHERE user_id=$1 AND product_id=$2 AND quantity=$3 AND created_at=$4;' with parameters: ['f9074cc1-fa47-429d-a66b-8a82e6c14b45', '20ff98cb-9b1f-4568-bd55-4190d99eb6b6', '1', '2026-04-19 17:59:37.2401304 -0400 EDT m=+0.214377701']
executing query 2: 'DELETE FROM products WHERE id=$1 AND company_id=$2 AND item_name=$3 AND price=$4 AND quantity=$5 AND description=$6 AND created_at=$7;' with parameters: ['20ff98cb-9b1f-4568-bd55-4190d99eb6b6', '45fdcb8a-dba6-44e6-b5f2-70811ba7ebf6', 'yiuFpH', '5.038534828802739', '1', 'QsTnSCDzln', '2026-04-19 17:59:37.2401304 -0400 EDT m=+0.214377701']
executing query 3: 'DELETE FROM companies WHERE created_at=$1 AND id=$2 AND name=$3 AND email=$4;' with parameters: ['2026-04-19 17:59:37.2401304 -0400 EDT m=+0.214377701', '45fdcb8a-dba6-44e6-b5f2-70811ba7ebf6', 'iLnnsgWNsD', 'vWFzvKGBmz']
executing query 4: 'DELETE FROM users WHERE id=$1 AND first_name=$2 AND last_name=$3 AND email=$4 AND address=$5 AND created_at=$6;' with parameters: ['f9074cc1-fa47-429d-a66b-8a82e6c14b45', 'BjIcVGfSKv', 'SzsPEPXOMh', 'gerGVBYVAb', 'exzCOeCLqg', '2026-04-19 17:59:37.2401304 -0400 EDT m=+0.214377701']      
Finished DELETE query execution! Clean up successful
```

# CLI Usage Details

The CLI has 3 subcommand palettes. These are `insert`, `format` and `validate`. `insert` focuses on
generating table entries. `format` focuses on creating, updating and verifying JSON template files. 
Finally, `validate` focuses on cycle detection, suggestion and resolution.

## Subcommand Palette: insert
This subcommand palette focuses on generating data. This data generation is either generating table entries
or outputting a set of INSERT and DElETE queries as files. The only command in the palette is the `entry` command

### entry
```
Command that is used to insert table entries into the database.
The user chooses if the table entries are generated from the default configuration
or from a specified template file. The command then goes on to produce INSERT queries
to add data to the database. If the user set the --clean-up flag to true, the command
will also generate the subsequent DELETE commands to cleanly remove only the data that
was generated by it. The delete commands will be ran after the command prompts the user via terminal.


examples:
        dbvg insert entry --database "$URL" --default --table "purchases" --verbose
        dbvg insert entry --database "$URL" --template "./templates/purchase_template.json" --table "purchases" --amount 10 -v --clean-up

Usage:
  dbvg insert entry [flags]

Flags:
  -c, --clean-up   cleans up after generating data
  -h, --help       help for entry
  -v, --verbose    Shows which queries are run and in what order

Global Flags:
      --amount int        amount of items to insert (default 1)
      --database string   url to connect to the database with
      --default           flag that determines if the default configuration is used
      --table string      name of table in the database
      --template string   path to the template file
```

### queries  [DEPRECATED as of v1.5.0]
```
Command that saves the generated queries to output files. These output
files are meant to provide the user the option to reuse generated queries. The queries are split between two files.
The INSERT queries are saved to a file with the extension .build.sql and the DELETE queries are saved to a
file with the extension .clean.sql

examples:
        dbvg generate queries --database "${URL}" --dir queries --amount 1 --template ./templates/purchase_template.json --table "purchases" --name "purchases"
        dbvg generate queries --database "${URL}" --dir queries/ --amount 1 --default --table "b" --name "purchases"

Usage:
  dbvg generate queries [flags]

Flags:
      --dir string    Path to the directory for the file output (default "./")
  -h, --help          help for queries
      --name string   Name of the output files

Global Flags:
      --amount int        amount of items to generate (default 1)
      --database string   url to connect to the database with
      --default           flag that determines if the default configuration is used
      --table string      name of sql table in the database
      --template string   path to the template file
```

## Subcommand Palette: format
This subcommand palette focuses on creating, updating and verifying template files. These templates are formatted as JSON files.
Templates are made to allow the user to control the behavior of the CLI and library. For more information
regarding templates, please read [the readme for the format sub palette](format/README.md). The only command in this palette is the `insert-template` command.

### insert-template
```
Command used to create and update insert templates which can be
used to customize the data generated using 'generate entry'. These templates can be updated
in the case where schema changes were made without losing relevant data. This command also allows
for template verification

ex)
        dbvg format insert-template --database "$URL" --create --path "some_path.json" -t "purchases"
        dbvg format insert-template --database "$URL" --update --path "some_path.json" -t "purchases"
        dbvg format insert-template --database "$URL" --verify --path "some_path.json" -t "purchases"

Usage:
  dbvg format insert-template [flags]

Flags:
  -c, --create         create a new template
  -h, --help           help for insert-template
  -t, --table string   the name of the table you want to create an insert template for
  -u, --update         update the given template with current schema information
      --verify         run deep verification on the template by checking codes and values

Global Flags:
      --database string   url to connect to the database with
  -p, --path string       specifies which file to use or where to output the template
```

## Subcommand Palette: validate
This subcommand palette focuses on schema validation. Specifically, detecting and resolving cyclic relationships between tables.
This palette has two commands, `schema` and `table`. 

### schema
```
Command used to validate the database schema and identify cycles. 
These cycles can immediately be resolved by running a set of queries. The user
has the option of simply viewing these suggestions or directly running them.

examples:
        dbvg validate schema --database "$URL" --run
        dbvg validate schema --database "$URL" --suggestions -v
        dbvg validate schema --database "$URL" -s -o "script.sql"

Usage:
  dbvg validate schema [flags]

Flags:
  -e, --end-early       exit with exit code 1 when a cycle is detected
  -f, --force           skip asking for verification and begin running the queries
  -h, --help            help for schema
  -o, --output string   output file name
  -r, --run             run suggestions queries
  -s, --suggestions     show suggestion queries
  -v, --verbose         verbose output

Global Flags:
      --database string   url to connect to the database with
```

### table

```
Command used to check if the given table is involved in any cycles. The command uses
DFS for cycle detection and will ignore any cycles that does not involve the given table.
This command will return a formatted string with the result of the process.

examples:
        dbvg validate table --database "$URL" --name "users" --run -v
        dbvg validate table --database "$URL" --name "users" --suggestions -v
        dbvg validate table --database "$URL" --name "users" -s -o "script.sql"

Usage:
  dbvg validate table [flags]

Flags:
  -e, --end-early       exit with exit code 1 when a cycle is detected
  -f, --force           skip asking for verification and begin running the queries
  -h, --help            help for table
  -n, --name string     name of the table in database
  -o, --output string   output file name
  -r, --run             run suggestions queries
  -s, --suggestions     show suggestion queries
  -v, --verbose         verbose output

Global Flags:
      --database string   url to connect to the database with
```


