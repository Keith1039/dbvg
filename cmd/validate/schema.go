package validate

import (
	"fmt"
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/graph"
	"github.com/Keith1039/dbvg/utils"
	"github.com/spf13/cobra"
	"log"
)

var (
	run         bool
	suggestions bool
	verbose     bool
	output      string
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
			if verbose { // only print each individual cycle if verbose is specified
				for _, cycle := range cycles {
					fmt.Println(fmt.Sprintf("Cycle Detected!: %s", cycle))
				}
			}
			fmt.Println(fmt.Sprintf("%d cycles detected", len(cycles)))
		} else {
			fmt.Println("No cycles detected!")
		}
		suggestionQueries := ordering.GetSuggestionQueries()
		if len(suggestionQueries) == 0 { // print that there's nothing to do
			fmt.Println("No suggestions to be made")
		}
		if suggestions {
			if output != "" { // write queries to file if there is one specified
				err = utils.WriteQueriesToFile(output, suggestionQueries)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				for i, query := range suggestionQueries { // print out each query
					fmt.Println(fmt.Sprintf("Query %d: %s", i+1, query))
				}
			}
		} else if run {
			if verbose { // only print each query if verbose is specified
				err = database.RunQueriesVerbose(db, suggestionQueries)
				if err != nil {
					log.Fatal(err)
				}
			} else { // run silently
				err := database.RunQueries(db, suggestionQueries)
				if err != nil {
					log.Fatal(err)
				}
			}
			fmt.Println("Queries ran successfully")
		}

	},
}

func init() {
	schemaCmd.Flags().BoolVarP(&run, "run", "r", false, "run suggestions queries")
	schemaCmd.Flags().BoolVarP(&suggestions, "suggestions", "s", false, "show suggestion queries")
	schemaCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	schemaCmd.Flags().StringVarP(&output, "output", "o", "", "output to specified file")
	schemaCmd.MarkFlagsMutuallyExclusive("suggestions", "run") // either you want the suggestions or you run them
	schemaCmd.MarkFlagsMutuallyExclusive("run", "output")      // if you're running the queries there's no need to output them

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// schemaCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// schemaCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
