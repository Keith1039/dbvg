/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package generate

import (
	"fmt"
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/parameters"
	"github.com/Keith1039/dbvg/utils"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
)

var (
	dir  string
	name string
)

// queriesCmd represents the queries command
var queriesCmd = &cobra.Command{
	Use:   "queries",
	Short: "Command that saves the generated queries to output files.",
	Long: `Command that saves the generated queries to output files. These output
files are meant to provide the user the option to reuse generated queries rather than
having to use the entry command to make them again. The commands are split between two files.
The INSERT queries are saved to a file with the extension .build.sql and the DELETE queries are saved to a 
file with the extension .clean.sql

examples:
	dbvg generate queries --database "${URL}" --dir something/somewhere --amount 500 --template some/file.json --table "b" --name "test"
	dbvg generate queries --database "${URL}" --dir something/somewhere --amount 500 --default --table "b" --name "test"
`,
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
		name = strings.TrimSpace(name)
		if name == "" {
			filePrefix = table
		} else {
			filePrefix = name
		}

		// if the folder doesn't exist, make it
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				log.Fatal(err)
			}
		}
		folder := strings.TrimSpace(dir) // trim space
		if folder[len(folder)-1] != '/' {
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
	queriesCmd.Flags().StringVarP(&dir, "dir", "", "./", "Path to the directory for the file output")
	queriesCmd.Flags().StringVarP(&name, "name", "", "", "Name of the output files")

	err := queriesCmd.MarkFlagDirname("dir") // mark the flag as a directory for autocomplete
	if err != nil {
		log.Fatal(err)
	}
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// queriesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// queriesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
