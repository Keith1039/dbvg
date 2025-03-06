package template

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
	dir  string
	name string
)

// createCmd represents the template command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Command used to generate a template in a specific folder for a group of tables",
	Long: `Command used to generate a template JSON file in a specific folder for a group of tables. 
The group of tables are the given table and all the tables it depends on.

examples:
	dbvg template create --database ${POSTGRES_URL} --dir "templates/"  --table "purchases"
	dbvg template create --database ${POSTGRES_URL} --dir "./templates/"  --table "purchases" --name "purchase_template.json"
`,
	Run: func(cmd *cobra.Command, args []string) {
		var filename string
		db, err := database.InitDB(ConnString) // starts up the database connection
		defer database.CloseDB(db)             // closes the database connection
		if err != nil {
			log.Fatal(err)
		}
		ordering := graph.NewOrdering(db)
		table = utils.TrimAndLowerString(table)
		tableOrder, err := ordering.GetOrder(table)
		if err != nil {
			log.Fatal(err)
		}

		templates := utils.MakeTemplates(db, tableOrder)
		jsonString, err := json.MarshalIndent(templates, "", "  ")
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, os.ModePerm)
			if err != nil {
				log.Fatal(err)
			}
		}

		folder := strings.TrimSpace(dir)  // trim space
		if folder[len(folder)-1] != '/' { // add closing backslash
			folder = folder + "/"
		}
		name = strings.TrimSpace(name) // trim space
		// check for empty string
		if name == "" {
			filename = fmt.Sprintf("%s_template.json", table)
		} else {
			filename = name
		}
		err = os.WriteFile(fmt.Sprintf("%s%s", folder, filename), jsonString, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

	},
}

func init() {
	createCmd.Flags().StringVarP(&dir, "dir", "", "./", "path to the output directory")
	createCmd.Flags().StringVarP(&name, "name", "", "", "name of the output template file")

	err := createCmd.MarkFlagDirname("dir") // mark it as a directory name
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
