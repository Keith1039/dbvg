package validate

import (
	"fmt"
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/graph"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var tableName string

// tableCmd represents the table command
var tableCmd = &cobra.Command{
	Use:   "table",
	Short: "Command used to check if the given table is involved in any cycles",
	Long: `Command used to check if the given table is involved in any cycles. The command uses
DFS for cycle detection and will ignore any cycles that does not involve the given table. 
This command will return a formatted string with the result of the process.

examples:
	dbvg validate table --database ${POSTGRES_URL} --name "users" --run -v
	dbvg validate table --database ${POSTGRES_URL} --name "users" --suggestions -v
	dbvg validate table --database ${POSTGRES_URL} --name "users" -s -o "script.sql"
`,
	Run: func(cmd *cobra.Command, args []string) {
		var cycles []string
		db, err := database.InitDB(ConnString) // starts up the database connection
		defer database.CloseDB(db)             // closes the database connection
		// error check
		if err != nil {
			log.Fatal(err)
		}
		ord := graph.NewOrdering(db)                   // get the ordering struct
		cycles, err = ord.GetCyclesForTable(tableName) // get the relevant cycles
		// error check
		if err != nil {
			log.Fatal(err)
		}
		// check if there are cycles and print the following
		size := len(cycles)
		if size > 0 {
			if verbose { // only print out the individual cycles if verbose is true
				fmt.Printf("The table '%s' is involved in %d cycles: \n%s", tableName, size, strings.Join(cycles, "\n"))
			} else {
				fmt.Printf("The table '%s' is involved in %d cycles", tableName, size)
			}
		} else {
			fmt.Printf("The table '%s' is not involved in any cycles.", tableName)
		}
		handleCmdFlags(db, ord, cycles) // handles the other flags
	},
}

func init() {
	// the name of the table being validated
	addFlags(tableCmd) // add the basic flags and their logic
	tableCmd.Flags().StringVarP(&tableName, "name", "n", "", "name of the table in database")
	err := tableCmd.MarkFlagRequired("name") // make name required
	// error check
	if err != nil {
		log.Fatal(err)
	}

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// tableCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// tableCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
