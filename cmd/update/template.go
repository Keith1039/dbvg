/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package update

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
	templatePath string
	tableName    string
)

// updateCmd represents the update command
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Command that updates an existing template",
	Long: `Command that updates an existing template. The command verifies for corruption before overwriting the current template with the new one.
This command also maps entries from the old template over to the new template.`,
	Run: func(cmd *cobra.Command, args []string) {
		// check to see if file exists
		if _, err := os.Stat(templatePath); !os.IsNotExist(err) {
			db, err := database.InitDB(ConnString) // start up the database
			defer database.CloseDB(db)             // close the database connection

			if err != nil {
				log.Fatal(err)
			}

			oldTemplate, err := verifyTemplate(templatePath) // verify that the old template is a valid template and return information
			if err != nil {
				log.Fatal(err)
			}
			ord := graph.NewOrdering(db)               // get a new ordering
			tableOrder, err := ord.GetOrder(tableName) // get the order of the tables
			if err != nil {
				log.Fatal(err)
			}

			newTemplate := utils.MakeTemplates(db, tableOrder)          // get a new blank template
			updateTemplate(oldTemplate, newTemplate)                    // update the new template with the info in the old template
			jsonBytes, err := json.MarshalIndent(newTemplate, "", "  ") // marshall the map
			if err != nil {
				log.Fatal(err)
			}

			err = os.WriteFile(templatePath, jsonBytes, os.ModePerm)
			if err != nil {
				log.Fatal(err)
			}

		} else {
			log.Fatal("incorrect path to template. Please verify that the file exists")
		}
	},
}

func init() {
	templateCmd.Flags().StringVarP(&templatePath, "path", "", "", "path to the template path")
	templateCmd.Flags().StringVarP(&tableName, "table", "", "", "name of the table that the template was generated from")
	err := templateCmd.MarkFlagRequired("path")
	if err != nil {
		log.Fatal(err)
	}
	err = templateCmd.MarkFlagRequired("table")
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

func verifyTemplate(templatePath string) (map[string]map[string]map[string]string, error) {
	m := make(map[string]map[string]map[string]string)
	bytes, err := os.ReadFile(templatePath) // read the bytes
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &m) // unmarshal the JSON
	if err != nil {
		return nil, err
	}

	// check the keys (doesn't verify the tables or columns yet)
	for _, columns := range m {
		for _, columnFields := range columns {
			_, ok := columnFields["Code"]
			if !ok {
				return nil, errors.New("corrupted template file detected")
			}
			_, ok = columnFields["Type"]
			if !ok {
				return nil, errors.New("corrupted template file detected")
			}
			_, ok = columnFields["Value"]
			if !ok {
				return nil, errors.New("corrupted template file detected")
			}
		}
	}
	return m, nil
}

func updateTemplate(oldTemplate map[string]map[string]map[string]string, newTemplate map[string]map[string]map[string]string) {
	for table, columns := range newTemplate {
		for columnName := range columns {
			_, ok := oldTemplate[table][columnName]
			if ok {
				newTemplate[table][columnName]["Code"] = oldTemplate[table][columnName]["Code"]   // set the code to the existing code
				newTemplate[table][columnName]["Value"] = oldTemplate[table][columnName]["Value"] // set the value to the existing value
			}
		}
	}
}
