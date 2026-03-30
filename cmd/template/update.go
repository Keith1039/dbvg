/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package template

import (
	"encoding/json"
	"errors"
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/graph"
	"github.com/Keith1039/dbvg/utils"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var (
	template string
)

// updateCmd represents the template command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Command that updates an existing template",
	Long: `Command that updates an existing template. The command verifies for file corruption, whether the file is formatted correctly, before overwriting 
the current template with the new one. This command also maps entries from the old template over to the new template, saving previous settings.

example:
	dbvg template update --database ${POSTGRES_URL} --template ./templates/purchase_template.json  --table "shop"
`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.InitDB(ConnString) // start up the database
		if err != nil {
			log.Fatal(err)
		}
		defer database.CloseDB(db) // close the database connection

		// check to see if file exists
		oldTemplate, err := utils.RetrieveInsertTemplateJSON(template) // verify that the old template is a valid template and return information
		if err != nil {
			log.Fatal(err)
		}
		err = verifyTemplate(oldTemplate) // shallow verification
		if err != nil {
			log.Fatal(err)
		}
		ord, err := graph.NewOrdering(db) // get a new ordering
		if err != nil {
			log.Fatal(err)
		}
		table = utils.TrimAndLowerString(table) // clean the table value
		tableOrder, err := ord.GetOrder(table)  // get the order of the tables
		if err != nil {
			log.Fatal(err)
		}

		newTemplate := utils.MakeTemplates(db, tableOrder)          // get a new blank template
		updateTemplate(oldTemplate, newTemplate)                    // template the new template with the info in the old template
		jsonBytes, err := json.MarshalIndent(newTemplate, "", "  ") // marshall the map
		if err != nil {
			log.Fatal(err)
		}

		err = os.WriteFile(template, jsonBytes, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

	},
}

func init() {
	updateCmd.Flags().StringVarP(&template, "template", "", "", "path to the template path")
	err := updateCmd.MarkFlagRequired("template") // mark it as required
	if err != nil {
		log.Fatal(err)
	}

	err = updateCmd.MarkFlagFilename("template") // make autocomplete look for files
	if err != nil {
		log.Fatal(err)
	}
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func verifyTemplate(m map[string]map[string]map[string]any) error {
	// check the keys (doesn't verify the tables or columns yet)
	for _, columns := range m {
		for _, columnFields := range columns {
			_, ok := columnFields["code"]
			if !ok {
				return errors.New("corrupted template file detected")
			}
			_, ok = columnFields["type"]
			if !ok {
				return errors.New("corrupted template file detected")
			}
			_, ok = columnFields["value"]
			if !ok {
				return errors.New("corrupted template file detected")
			}
		}
	}
	return nil
}

func updateTemplate(oldTemplate map[string]map[string]map[string]any, newTemplate map[string]map[string]map[string]any) {
	for tableName, columns := range newTemplate {
		for columnName := range columns {
			_, ok := oldTemplate[tableName][columnName]
			if ok {
				newTemplate[tableName][columnName]["code"] = oldTemplate[tableName][columnName]["code"]   // set the code to the existing code
				newTemplate[tableName][columnName]["value"] = oldTemplate[tableName][columnName]["value"] // set the value to the existing value
			}
		}
	}
}
