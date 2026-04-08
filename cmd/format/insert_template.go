/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package format

import (
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/graph"
	"github.com/Keith1039/dbvg/template"
	"github.com/Keith1039/dbvg/utils"
	"github.com/spf13/cobra"
	"log"
)

var table string
var update bool
var verify bool

// insertTemplateCmd represents the insertTemplate command
var insertTemplateCmd = &cobra.Command{
	Use:   "insert-template",
	Short: "Command used to create and update insert templates.",
	Long: `Command used to create and update insert templates which can be
used to customize the data generated via other commands. These templates can be updated
in the case where schema changes were made without losing relevant data.

ex)
	dbvg format insert-template --database "$URL" --path "some_path.json" -t "purchases"
	dbvg format insert-template --database "$URL" -u --path "some_path.json" -t "purchases"
	dbvg format insert-template --database "$URL" --verify --path "some_path.json" -t "purchases"
`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.InitDB(ConnString) // starts up the database connection
		if err != nil {
			log.Fatal(err)
		}
		defer database.CloseDB(db) // closes the database connection
		ordering, err := graph.NewOrdering(db)
		if err != nil {
			log.Fatal(err)
		}
		table = utils.TrimAndLowerString(table)
		tableOrder, err := ordering.GetOrder(table)
		if err != nil {
			log.Fatal(err)
		}
		templates := utils.MakeTemplates(db, tableOrder)
		if update {
			err = utils.UpdateInsertTemplate(path, templates)
			if err != nil {
				log.Fatal(err)
			}
		} else if verify {
			_, err = template.NewInsertTemplate(database.GetAllColumnData(db), tableOrder, path)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			err = utils.WriteInsertTemplateToFile(path, templates)
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}

func init() {
	insertTemplateCmd.Flags().StringVarP(&table, "table", "t", "", "the name of the table you want to create an insert template for")
	insertTemplateCmd.Flags().BoolVarP(&update, "update", "u", false, "update the table if they already exist")
	insertTemplateCmd.Flags().BoolVarP(&verify, "verify", "", false, "provide deep verification on the template")

	err := insertTemplateCmd.MarkFlagRequired("table")
	if err != nil {
		log.Fatal(err)
	}
	insertTemplateCmd.MarkFlagsMutuallyExclusive("update", "verify")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// insertTemplateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// insertTemplateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
