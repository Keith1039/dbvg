package validate

import (
	"fmt"
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/graph"
	"github.com/spf13/cobra"
	"log"
)

var (
	run         bool
	suggestions bool
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
	dbvg validate schema --database ${POSTGRES_URL} --suggestions
`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.InitDB(ConnString)
		defer database.CloseDB(db)

		if err != nil {
			log.Fatal(err)
		}
		ordering := graph.NewOrdering(db)
		cycles := ordering.GetCycles()
		if len(cycles) > 0 {
			for _, cycle := range cycles {
				fmt.Println(fmt.Sprintf("Cycle Detected!: %s", cycle))
			}
		} else {
			fmt.Println("No cycles detected!")
		}
		if suggestions {
			suggestions := ordering.GetSuggestionQueries()
			if len(suggestions) > 0 {
				for i, query := range suggestions {
					fmt.Println(fmt.Sprintf("Query %d: %s", i+1, query))
				}
			} else {
				fmt.Println("No suggestions to be made")
			}
		} else if run {
			suggestions := ordering.GetSuggestionQueries()
			if len(suggestions) > 0 {
				for i, query := range suggestions {
					fmt.Println(fmt.Sprintf("Query %d: %s", i+1, query))
				}
				err := database.RunQueries(db, suggestions)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println("Queries ran successfully")
			} else {
				fmt.Println("No suggestions to be made")
			}
		}

	},
}

func init() {
	schemaCmd.Flags().BoolVarP(&run, "run", "r", false, "run suggestions queries")
	schemaCmd.Flags().BoolVarP(&suggestions, "suggestions", "s", false, "show suggestion queries")
	schemaCmd.MarkFlagsMutuallyExclusive("suggestions", "run")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// schemaCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// schemaCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
