package generate

import (
	"bufio"
	"database/sql"
	"fmt"
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/parameters"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var (
	table         string
	template      string
	amount        int
	verbose       bool
	cleanUp       bool
	defaultConfig bool
)

// entryCmd represents the entry command
var entryCmd = &cobra.Command{
	Use:   "entry",
	Short: "Command used to generate table entries",
	Long: `Command that is used to generate table entries in the database.
	This command requires the --table flag and either the --template or --default flags.
	If you want the entries to disappear after execution use the --clean-up flag.
	You can control how many entries generated with the --amount flag (default is 1).
	Finally, if you want more information regarding the execution use -v or --verbose for a more verbose output.

	examples of valid commands)
		dbvg generate entry --database ${POSTGRES_URL} --default --table "example_table" --verbose
		dbvg generate entry --database ${POSTGRES_URL} --template "path/to/file.json" --table "example_table" --amount 10 -v --clean-up
	`,
	Run: func(cmd *cobra.Command, args []string) {

		var writer *parameters.QueryWriter
		db, err := InitDB()
		defer func(db *sql.DB) {
			err := db.Close()
			if err != nil {
				log.Fatal(err)
			}
		}(db)
		if err != nil {
			log.Fatal(err)
		}
		tMap := database.GetTableMap(db)
		_, ok := tMap[table]
		if !ok {
			log.Fatalf("Table %s does not exist in database", table)
		}
		if amount <= 0 {
			log.Fatal("amount must be greater than zero")
		}
		if defaultConfig {
			writer, err = parameters.NewQueryWriterFor(db, table)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			writer, err = parameters.NewQueryWriterWithTemplateFor(db, table, template)
			if err != nil {
				log.Fatal(err)
			}
		}
		insertQueries, deleteQueries := writer.GenerateEntries(amount)

		fmt.Println("Beginning INSERT query execution...")
		if verbose {
			err = database.RunQueriesVerbose(db, insertQueries)
		} else {
			err = database.RunQueries(db, insertQueries)
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Finished INSERT query execution!")
		if cleanUp {
			br := bufio.NewReader(os.Stdin)
			fmt.Print("Press Enter to begin clean up: ")
			br.ReadString('\n') // error doesn't matter
			fmt.Println("Beginning DELETE query execution...")
			if verbose {
				err = database.RunQueriesVerbose(db, deleteQueries)
			} else {
				err = database.RunQueries(db, deleteQueries)
			}
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Finished DELETE query execution! Clean up successful")
		}
	},
}

func init() {
	entryCmd.Flags().StringVarP(&template, "template", "", "", "path to the template file being used")
	entryCmd.Flags().StringVarP(&table, "table", "", "", "table we are generating data for")
	entryCmd.Flags().IntVarP(&amount, "amount", "", 1, "amount of entries this will generate")
	entryCmd.Flags().BoolVarP(&defaultConfig, "default", "", false, "run using the default template")
	entryCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Shows which queries are run and in what order")
	entryCmd.Flags().BoolVarP(&cleanUp, "clean-up", "", false, "cleans up after generating data")
	err := entryCmd.MarkFlagRequired("table")
	if err != nil {
		log.Fatal(err)
	}
	entryCmd.MarkFlagsOneRequired("template", "default")
	entryCmd.MarkFlagsMutuallyExclusive("template", "default") // either use a template or use the default

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// entryCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// entryCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
