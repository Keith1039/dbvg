package generate

import (
	"bufio"
	"fmt"
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/parameters"
	"github.com/Keith1039/dbvg/utils"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var (
	verbose bool
	cleanUp bool
)

// entryCmd represents the entry command
var entryCmd = &cobra.Command{
	Use:   "entry",
	Short: "Command used to generate table entries",
	Long: `Command that is used to generate table entries in the database.
The user chooses if the table entries are generated from the default configuration
or from a specified template file.

examples:
	dbvg generate entry --database ${POSTGRES_URL} --default --table "purchases" --verbose
	dbvg generate entry --database ${POSTGRES_URL} --template "./templates/purchase_template.json" --table "purchases" --amount 10 -v --clean-up
`,
	Run: func(cmd *cobra.Command, args []string) {

		var writer *parameters.QueryWriter
		db, err := database.InitDB(ConnString) // starts up database connection
		defer database.CloseDB(db)             // closes the database connection

		if err != nil {
			log.Fatal(err)
		}
		tMap := database.GetTableMap(db)
		table = utils.TrimAndLowerString(table)
		_, ok := tMap[table]
		if !ok {
			log.Fatalf("Table %s does not exist in database", table)
		}
		if amount <= 0 {
			log.Fatal("amount must be greater than zero")
		}
		if defaultConfig {
			writer, err = parameters.NewQueryWriter(db, table)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			writer, err = parameters.NewQueryWriterWithTemplate(db, table, template)
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
	entryCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Shows which queries are run and in what order")
	entryCmd.Flags().BoolVarP(&cleanUp, "clean-up", "c", false, "cleans up after generating data")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// entryCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// entryCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
