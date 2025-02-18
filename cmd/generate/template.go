package generate

import (
	"container/list"
	"database/sql"
	"encoding/json"
	"fmt"
	database "github.com/Keith1039/dbvg/db"
	"github.com/Keith1039/dbvg/graph"
	"github.com/jimsmart/schema"
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
	Short: "generates a template in a specific folder for a specific group of tables",
	Long: `generates a template JSON file in a specific folder for a specific group of tables based off of the first
	table given. This template is meant to be edited by the user and ingested by either the CLI or the library. As a result,
	the --dir and --table flags are required.

	example of valid command)
		dbvg generate template --database ${POSTGRES_URL} --dir "some/directory"  --table "example_table"
	`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := InitDB()
		if err != nil {
			log.Fatal(err)
		}
		ordering := graph.NewOrdering(db)

		tableOrder, err := ordering.GetOrder(strings.ToLower(tableName))
		if err != nil {
			log.Fatal(err)
		}
		templates := makeTemplates(db, tableOrder)
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

func makeTemplates(db *sql.DB, l *list.List) map[string]map[string]map[string]string {
	m := make(map[string]map[string]map[string]string)
	relations := database.GetRelationships(db) // get relationships
	node := l.Front()
	for node != nil {
		tName := node.Value.(string)
		m[tName] = makeTemplate(db, tName, relations)
		node = node.Next()
	}
	return m
}

func makeTemplate(db *sql.DB, tName string, relations map[string]map[string]map[string]string) map[string]map[string]string {
	m := make(map[string]map[string]string)
	cols, err := schema.ColumnTypes(db, "", tName)
	colMap := database.GetColumnMap(db, tName)
	if err != nil {
		log.Fatal(err)
	}
	for _, col := range cols {
		_, ok := relations[tName][col.Name()] // check if the column is a fk
		if !ok {
			m[col.Name()] = map[string]string{"Type": colMap[col.Name()], "Code": "", "Value": ""}
		}
	}
	return m
}
