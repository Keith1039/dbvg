# Common Uses

## Validate your database Schema using CLI

`dbvg validate schema --database ${POSTGRES_URL}`

Sample output:
```
Cycle Detected!: b --> c --> a --> b
Cycle Detected!: b --> d --> e --> b
```

## Resolve cycles using CLI

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

## Generate a Template
`dbvg generate template --database ${POSTGRES_URL} --dir "dir/" --table "b"`

For more regarding templates, please see [this](generate/README.md)

## Generate a table entry
`dbvg generate template --database ${POSTGRES_URL} --default --table "b" -v`

sample output:
```
Query 1: INSERT INTO a (akey) VALUES (0);
Query 2: INSERT INTO c (ckey, aref) VALUES (0, 0);
Query 3: INSERT INTO e (ekey) VALUES (0);
Query 4: INSERT INTO d (dkey, eref) VALUES (0, 0);
Query 5: INSERT INTO b (bkey2, cref, dref, bkey) VALUES (0, 0, 0, 0);
Finished INSERT query execution!
```

# CLI Usage Details

The CLI comprises the 2 subcommand palettes. These are `validate` and `generate`, `validate` is focused on database schema
validation and `generate` is focused on database table entry creation.

*Note*: In all the examples, for the connection string I use an environmental variable, ${POSTGRES_URL}. That is
because connections strings tend to be very long, and it would make the commands look less concise.
You can represent the connection string any way you like so long as it's value is a valid sql connection string.

## Subcommand Palette: validate
This subcommand palette is focused on schema validation. This means, detecting and resolving database cyclic relationships.
As such, this palette only has one command, that being `schema`. 

### schema
```
command used to validate the database schema and identify cycles. These
cycles can be resolved immediately by using the --run flag or the suggestions can be
printed without running them by using the --suggestions flag. These two flags cannot be
used simultaneously

example of valid commands)
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

## Subcommand Palette: generate
This subcommand palette is focused on generating data. This data generation is either generating a template for future use
or database table entries. As such, the two commands in the palette are `template` and `entry`

### template

```
generates a template JSON file in a specific folder for a specific group of tables based off of the first
table given. This template is meant to be edited by the user and ingested by either the CLI or the library. As a result,
the --dir and --table flags are required.

example of valid command)
        dbvg generate template --database ${POSTGRES_URL} --dir "some/directory"  --table "example_table"

Usage:
  dbvg generate template [flags]

Flags:
      --dir string     relative path of a directory to place the template file in, if the path doesn't exist it will make the folder
  -h, --help           help for template
      --table string   the name of the table we want an entry for

Global Flags:
      --database string   url to connect to the database with

```

### entry

```
Command that is used to generate table entries in the database.
This command requires the --table flag and either the --template or --default flags.
If you want the entries to disappear after execution use the --clean-up flag.
You can control how many entries generated with the --amount flag (default is 1).
Finally, if you want more information regarding the execution use -v or --verbose for a more verbose output.

examples of valid commands)
        dbvg generate entry --database ${POSTGRES_URL} --default --table "example_table" --verbose
        dbvg generate entry --database ${POSTGRES_URL} --template "path/to/file.json" --table "example_table" --amount 10 -v --clean-up

Usage:
  dbvg generate entry [flags]

Flags:
      --amount int        amount of entries this will generate (default 1)
      --clean-up          cleans up after generating data
      --default           run using the default template
  -h, --help              help for entry
      --table string      table we are generating data for
      --template string   path to the template file being used
  -v, --verbose           Shows which queries are run and in what order

Global Flags:
      --database string   url to connect to the database with

```

