/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package format

import (
	"fmt"
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/graph"
	"github.com/Keith1039/dbvg/template"
	"github.com/Keith1039/dbvg/utils"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var table string
var update bool
var verify bool
var create bool

// insertTemplateCmd represents the insertTemplate command
var insertTemplateCmd = &cobra.Command{
	Use:   "insert-template",
	Short: "Command used to create and update insert templates.",
	Long: `Command used to create and update insert templates which can be
used to customize the data generated using 'generate entry'. These templates can be updated
in the case where schema changes were made without losing relevant data. This command also allows
for template verification

ex)
	dbvg format insert-template --database "$URL" --create --path "some_path.json" -t "purchases"
	dbvg format insert-template --database "$URL" --update --path "some_path.json" -t "purchases"
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
			changes, err := utils.UpdateInsertTemplate(path, templates)
			if err != nil {
				log.Fatal(err)
			}
			if len(changes) > 0 {
				fmt.Println(fmt.Sprintf("the following changes were applied to the template at path '%s':", path))
				fmt.Println(strings.Join(changes, "\n"))
			} else {
				fmt.Println(fmt.Sprintf("no changes made to the template at path '%s'", path))
			}
		} else if verify {
			_, err = template.NewInsertTemplate(db, table, path)
			if err != nil {
				log.Fatalf("template at '%s' failed with error [%v]", path, err)
			} else {
				fmt.Println(fmt.Sprintf("template at '%s' contains no errors", path))
			}
		} else {
			// create code
			err = utils.WriteInsertTemplateToFile(path, templates)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(fmt.Sprintf("template successfully created at '%s'", path))
		}
	},
}

func init() {
	insertTemplateCmd.Flags().StringVarP(&table, "table", "t", "", "the name of the table you want to create an insert template for")
	insertTemplateCmd.Flags().BoolVarP(&update, "update", "u", false, "update the given template with current schema information")
	insertTemplateCmd.Flags().BoolVarP(&verify, "verify", "", false, "run deep verification on the template by checking codes and values")
	insertTemplateCmd.Flags().BoolVarP(&create, "create", "c", false, "create a new template")
	insertTemplateCmd.MarkFlagsOneRequired("create", "update", "verify")
	err := insertTemplateCmd.MarkFlagRequired("table")
	if err != nil {
		log.Fatal(err)
	}
	insertTemplateCmd.MarkFlagsMutuallyExclusive("update", "verify")
	insertTemplateCmd.MarkFlagsMutuallyExclusive("create", "verify")
	insertTemplateCmd.MarkFlagsMutuallyExclusive("create", "update")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// insertTemplateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// insertTemplateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
