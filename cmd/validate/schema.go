/*
Copyright © 2025 Keith Compere <KeithCompere150@gmail.com>
*/
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
	Short: "command used to validate the entire schema",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := InitDB()
		if err != nil {
			log.Fatal(err)
		}
		ordering := graph.NewOrdering(db)
		cycles := ordering.GetCycles()
		node := cycles.Front()
		if cycles.Len() > 0 {
			for node != nil {
				fmt.Println(fmt.Sprintf("Cycle Detected!: %s", node.Value.(string)))
				node = node.Next()
			}
		} else {
			fmt.Println("No cycles detected!")
		}
		if suggestions {
			suggestions := ordering.GetSuggestionQueries()
			if suggestions.Len() > 0 {
				node = suggestions.Front()
				i := 1
				for node != nil {
					fmt.Println(fmt.Sprintf("Query %d: %s", i, node.Value.(string)))
					node = node.Next()
					i++
				}
			} else {
				fmt.Println("No suggestions to be made")
			}
		} else if run {
			suggestions := ordering.GetSuggestionQueries()
			if suggestions.Len() > 0 {
				node = suggestions.Front()
				i := 1
				for node != nil {
					fmt.Println(fmt.Sprintf("Query %d: %s", i, node.Value.(string)))
					node = node.Next()
					i++
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
