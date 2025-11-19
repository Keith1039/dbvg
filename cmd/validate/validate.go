// Package validate provides the commands relating to schema validation
//
// This package contains the validation code for the CLI. It mirrors the functionality and depends on the `graph` package
package validate

import (
	"database/sql"
	"errors"
	"fmt"
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/graph"
	"github.com/Keith1039/dbvg/utils"
	"github.com/spf13/cobra"
	"log"
)

var (
	ConnString  string
	run         bool
	suggestions bool
	verbose     bool
	output      string
)

// ValidateCmd represents the validate command
var ValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "The palette responsible for schema validation.",
	Long: `This palette is responsible for detecting and removing cycles from the database schema.
Suggestions can also be given regarding the detected cycles.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func addSubCommands() {
	ValidateCmd.AddCommand(schemaCmd)
	ValidateCmd.AddCommand(tableCmd)
}

func init() {
	addSubCommands()
	ValidateCmd.PersistentFlags().StringVarP(&ConnString, "database", "", "", "url to connect to the database with")

	if err := ValidateCmd.MarkPersistentFlagRequired("database"); err != nil {
		log.Fatal(err)
	}

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// validateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// validateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func handleCmdFlags(db *sql.DB, ordering *graph.Ordering, cycles []string) {
	var err error
	if suggestions || run { // global variables
		// functional equivalent to calling `GetSuggestionQueries` in some cases
		suggestionQueries := ordering.GetSuggestionQueriesForCycles(cycles)
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
				err = database.RunQueries(db, suggestionQueries)
				if err != nil {
					log.Fatal(err)
				}
			}
			fmt.Println("Queries ran successfully")
		}
	}
}

func addFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&run, "run", "r", false, "run suggestions queries")
	cmd.Flags().BoolVarP(&suggestions, "suggestions", "s", false, "show suggestion queries")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file name")
	cmd.MarkFlagsMutuallyExclusive("suggestions", "run")          // either you want the suggestions or you run them
	cmd.PreRunE = func(cmd *cobra.Command, args []string) error { // define the pre-run function
		flagSet := cmd.Flags()                                            // get flag set
		if flagSet.Changed("output") && !flagSet.Changed("suggestions") { // check to see if output is set but suggestions isn't
			return errors.New("'suggestions' flag must be set for 'output' flag to be used") // format error
		}
		return nil
	}
	cmd.MarkFlagsMutuallyExclusive("run", "output") // if you're running the queries there's no need to output them
}
