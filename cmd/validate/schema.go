package validate

import (
	"fmt"
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/graph"
	"github.com/spf13/cobra"
	"log"
)

// schemaCmd represents the schema command
var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Command used to validate the entire database schema.",
	Long: `Command used to validate the database schema and identify cycles. 
These cycles can immediately be resolved by running a set of queries or
these suggestions to the user.

examples:
	dbvg validate schema --database ${POSTGRES_URL} --run
	dbvg validate schema --database ${POSTGRES_URL} --suggestions -v
	dbvg validate schema --database ${POSTGRES_URL} -s -o "script.sql"
`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.InitDB(ConnString)
		defer database.CloseDB(db)
		if err != nil {
			log.Fatal(err)
		}
		ord := graph.NewOrdering(db)
		cycles := ord.GetCycles()
		if len(cycles) > 0 {
			if verbose { // only print each individual cycle if verbose is specified
				for _, cycle := range cycles {
					fmt.Println(fmt.Sprintf("Cycle Detected!: %s", cycle))
				}
			}
			fmt.Println(fmt.Sprintf("%d cycles detected", len(cycles)))
		} else {
			fmt.Println("No cycles detected!")
		}
		handleCmdFlags(db, ord, cycles) // handles the flag logic
	},
}

func init() {
	addFlags(schemaCmd) // add the flags and their basic logic

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// schemaCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// schemaCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
