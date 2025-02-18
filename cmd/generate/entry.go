/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package generate

import (
	"bufio"
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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var writer *parameters.QueryWriter
		db, err := InitDB()
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
			writer.GenerateEntries(amount)
		} else {
			if _, err := os.Stat(template); !os.IsNotExist(err) {
				writer, err = parameters.NewQueryWriterWithTemplateFor(db, table, template)
				if err != nil {
					log.Fatal(err)
				}
				writer.GenerateEntries(amount)
			} else {
				log.Fatal("template path is invalid")
			}
		}
		fmt.Println("Beginning INSERT query execution...")
		if verbose {
			err = database.RunQueriesVerbose(db, writer.InsertQueryQueue)
		} else {
			err = database.RunQueries(db, writer.InsertQueryQueue)
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
				err = database.RunQueriesVerbose(db, writer.DeleteQueryQueue)
			} else {
				err = database.RunQueries(db, writer.DeleteQueryQueue)
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
