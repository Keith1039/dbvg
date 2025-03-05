/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package generate

import (
	"fmt"
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/parameters"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
)

var (
	amounts      int
	folder       string
	defaultR     bool
	templatePath string
	name         string
)

// queriesCmd represents the queries command
var queriesCmd = &cobra.Command{
	Use:   "queries",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		var writer *parameters.QueryWriter
		var filePrefix string
		var buildFile, cleanUpFile *os.File

		db, err := database.InitDB(ConnString) // starts up database connection
		defer database.CloseDB(db)             // closes the database connection
		defer buildFile.Close()                // close the build up file
		defer cleanUpFile.Close()              // close the cleanup file

		if err != nil {
			log.Fatal(err)
		}
		tMap := database.GetTableMap(db)
		table = strings.ToLower(table) // make it lower case
		_, ok := tMap[table]
		if !ok {
			log.Fatalf("Table %s does not exist in database", table)
		}
		if amount <= 0 {
			log.Fatal("amount must be greater than zero")
		}
		if defaultR {
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
		name = strings.TrimSpace(name)
		if name != "" {
			filePrefix = table
		} else {
			filePrefix = name
		}

		folder := strings.TrimSpace(folder) // trim space
		if folder[len(folder)-1:] != "/" {
			folder = folder + "/"
		}

		buildFile, err = os.Create(fmt.Sprintf("%s%s_query.build.sql", folder, filePrefix))
		if err != nil {
			log.Fatal(err)
		}
		cleanUpFile, err = os.Create(fmt.Sprintf("%s%s_query.clean.sql", folder, filePrefix))
		if err != nil {
			log.Fatal(err)
		}
		writeToFile(buildFile, insertQueries)
		writeToFile(cleanUpFile, deleteQueries)
	},
}

func writeToFile(file *os.File, queries []string) {
	for _, query := range queries {
		_, err := fmt.Fprintln(file, query)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func init() {
	queriesCmd.Flags().IntVarP(&amounts, "amount", "", 0, "Amount of queries to generate")
	queriesCmd.Flags().StringVarP(&folder, "dir", "", "./", "Path to the directory for the file output")
	queriesCmd.Flags().BoolVarP(&defaultR, "default", "", false, "Run with the default config")
	queriesCmd.Flags().StringVarP(&templatePath, "template", "", "", "Path to the template file")
	queriesCmd.Flags().StringVarP(&name, "name", "", "", "Name of the output files")

	entryCmd.MarkFlagsOneRequired("template", "default")
	entryCmd.MarkFlagsMutuallyExclusive("template", "default") // either use a template or use the default

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// queriesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// queriesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
