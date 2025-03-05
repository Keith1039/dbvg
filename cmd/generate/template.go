package generate

import (
	"encoding/json"
	"fmt"
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/graph"
	"github.com/Keith1039/dbvg/utils"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
)

var (
	dirPath   string
	tableName string
)

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Command used to generate a template in a specific folder, for a group of tables",
	Long: `Command used to generate a template JSON file in a specific folder, for a group of tables. 
The group of tables is based off of the first table given by the user. 
This template is meant to be edited by the user and ingested by either the CLI or the library. As a result,
the --dir and --table flags are required.

example:
	dbvg generate template --database ${POSTGRES_URL} --dir "some/directory"  --table "example_table"
`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.InitDB(ConnString) // starts up the database connection
		defer database.CloseDB(db)             // closes the database connection
		if err != nil {
			log.Fatal(err)
		}
		ordering := graph.NewOrdering(db)

		tableOrder, err := ordering.GetOrder(strings.ToLower(tableName))
		if err != nil {
			log.Fatal(err)
		}
		templates := utils.MakeTemplates(db, tableOrder)
		jsonString, err := json.MarshalIndent(templates, "", "  ")
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			err = os.MkdirAll(dirPath, os.ModePerm)
			if err != nil {
				log.Fatal(err)
			}
		}
		err = os.WriteFile(fmt.Sprintf("%s/%s_template.json", dirPath, tableName), jsonString, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

	},
}

func init() {

	templateCmd.Flags().StringVarP(&dirPath, "dir", "", "", "relative path of a directory to place the template file in, if the path doesn't exist it will make the folder")
	templateCmd.Flags().StringVarP(&tableName, "table", "", "", "the name of the table we want an entry for")

	err := templateCmd.MarkFlagRequired("dir")
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
	// templateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// templateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
