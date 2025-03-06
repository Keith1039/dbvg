# Common Uses

## Validate your database Schema using CLI

`dbvg validate schema --database ${POSTGRES_URL}`

Sample output:
```
Cycle Detected!: b --> c --> a --> b
Cycle Detected!: b --> d --> e --> b
```

## Resolve cycles using CLI [[schema used]](../db/migrations/case8/000001_create_compound_table.up.sql)
`dbvg validate schema --database ${POSTGRES_URL} --run`

Sample output:
```
Cycle Detected!: c --> a --> b --> c
Cycle Detected!: b --> d --> e --> b
Query 1: ALTER TABLE a DROP COLUMN bref;
Query 2: ALTER TABLE a DROP COLUMN bref2;
Query 3: CREATE TABLE IF NOT EXISTS b_a(
         b_bkey_ref INT4,
         b_bkey2_ref INT4,
         a_akey_ref INT4,
        FOREIGN KEY (b_bkey_ref, b_bkey2_ref) REFERENCES b(bkey, bkey2),
        FOREIGN KEY (a_akey_ref) REFERENCES a(akey),
        PRIMARY KEY (b_bkey_ref, b_bkey2_ref, a_akey_ref)
)
Query 4: ALTER TABLE e DROP COLUMN bref;
Query 5: ALTER TABLE e DROP COLUMN bref2;
Query 6: CREATE TABLE IF NOT EXISTS b_e(
         b_bkey_ref INT4,
         b_bkey2_ref INT4,
         e_ekey_ref INT4,
        FOREIGN KEY (b_bkey_ref, b_bkey2_ref) REFERENCES b(bkey, bkey2),
        FOREIGN KEY (e_ekey_ref) REFERENCES e(ekey),
        PRIMARY KEY (b_bkey_ref, b_bkey2_ref, e_ekey_ref)
)
Queries ran successfully
```

## Generate a template
`dbvg template create --database ${POSTGRES_URL} --dir "templates/" --table "purchases" --name "purchase_template.json"`

## Update an existing template
`dbvg template update --database ${POSTGRES_URL} --template ./templates/purchase_template.json  --table "purchases"`

For more regarding templates, please see [this](generate/README.md)

## Generate a table entry [[schema used]](../db/real_migrations/000001_shop_example.up.sql)
`dbvg generate entry --database ${POSTGRES_URL} --amount 1 --default --table "purchases" -v`

sample output:
```
Beginning INSERT query execution...
Query 1: INSERT INTO users (name, last_name, email, address, created, id) VALUES ('xODhzCQGqX', 'esvogNvCwh', 'vvmrQWrfwg', 'qBOXYSvuEM', '2025-03-06 18:17:23', '85c25f76-4e80-4bd2-98bf-439426f3ad23');
Query 2: INSERT INTO companies (email, created, id, name) VALUES ('BFngvsLunt', '2025-03-06 18:17:23', '9ef36a65-e7e0-4dc0-abff-fa62b7de6103', 'IbYlbEbpda');
Query 3: INSERT INTO products (created, id, company_id, item_name, price, quantity, description) VALUES ('2025-03-06 18:17:23', 'e32eb3b6-62f4-463d-aea2-a00002c5c578', '9ef36a65-e7e0-4dc0-abff-fa62b7de6103', 'WwJeJu', 38.86501279821432::MONEY, 1, 'bCUmKcDepV');
Query 4: INSERT INTO purchases (quantity, created, user_id, product_id) VALUES (1, '2025-03-06 18:17:23', '85c25f76-4e80-4bd2-98bf-439426f3ad23', 'e32eb3b6-62f4-463d-aea2-a00002c5c578');
Finished INSERT query execution!
```

### Generate INSERT and DELETE queries (with specified output file name)
`dbvg generate queries --database ${POSTGRES_URL} --amount 1 --default --table "purchases" --name "purchase_queries"`

# CLI Usage Details

The CLI has 3 subcommand palettes. These are `generate`, `template` and `validate`. `generate` focuses on
generating table entries. `Template` focuses on generating and updating JSON template files. 
Finally, `validate` focuses on cycle detection, suggestion and resolution.

*Note*: In all the examples, for the connection string I use an environmental variable, ${POSTGRES_URL}. That is
because connections strings tend to be very long, and it would make the commands look less concise.
You can represent the connection string any way you like so long as it is a valid sql connection string.

## Subcommand Palette: generate
This subcommand palette focuses on generating data. This data generation is either generating table entries
or outputting a set of INSERT and DElETE queries as files. The two commands in the palette are `entry` and `queries`.

### entry
```
Command that is used to generate table entries in the database.
The user chooses if the table entries are generated from the default configuration
or from a specified template file.

examples:
        dbvg generate entry --database ${POSTGRES_URL} --default --table "purchases" --verbose
        dbvg generate entry --database ${POSTGRES_URL} --template "./templates/purchase_template.json" --table "purchases" --amount 10 -v --clean-up

Usage:
  dbvg generate entry [flags]

Flags:
  -c, --clean-up   cleans up after generating data
  -h, --help       help for entry
  -v, --verbose    Shows which queries are run and in what order

Global Flags:
      --amount int        amount of items to generate (default 1)
      --database string   url to connect to the database with
      --default           flag that determines if the default configuration is used
      --table string      name of sql table in the database
      --template string   path to the template file
```

### queries
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

## Subcommand Palette: template
This subcommand palette focuses on generating and updating template files. These templates are formatted as JSON files.
Templates are made to allow the user to control what type of data is being generated through the CLI. For more information
regarding templates, please read [THIS](template/README.md). The commands in this palette are `create` and `update`.

### create
```
Command used to generate a template JSON file in a specific folder for a group of tables. 
The group of tables are the given table and all the tables it depends on.

examples:
        dbvg template create --database ${POSTGRES_URL} --dir "templates/"  --table "purchases"
        dbvg template create --database ${POSTGRES_URL} --dir "./templates/"  --table "purchases" --name "purchase_template.json"

Usage:
  dbvg template create [flags]

Flags:
      --dir string    path to the output directory (default "./")
  -h, --help          help for create
      --name string   name of the output template file

Global Flags:
      --database string   url to connect to the database with
      --table string      the name of the sql table that the template is based off of
```

### update
```
Command that updates an existing template. The command verifies for file corruption, whether the file is formatted correctly, before overwriting 
the current template with the new one. This command also maps entries from the old template over to the new template, saving previous settings.

example:
        dbvg template update --database ${POSTGRES_URL} --template ./templates/purchase_template.json  --table "shop"

Usage:
  dbvg template update [flags]

Flags:
  -h, --help              help for update
      --template string   path to the template path

Global Flags:
      --database string   url to connect to the database with
      --table string      the name of the sql table that the template is based off of
```

## Subcommand Palette: validate
This subcommand palette focuses on schema validation. Specifically, detecting and resolving cyclic relationships between tables.
This palette has one command, `schema`. 

### schema
```
Command used to validate the database schema and identify cycles. 
These cycles can immediately be resolved by running a set of queries or
these suggestions to the user.

examples:
        dbvg validate schema --database ${POSTGRES_URL} --run
        dbvg validate schema --database ${POSTGRES_URL} --suggestions

Usage:
  dbvg validate schema [flags]

Flags:
  -h, --help          help for schema
  -r, --run           run suggestions queries
  -s, --suggestions   show suggestion queries

Global Flags:
      --database string   url to connect to the database with
```


