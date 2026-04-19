// Package insert is responsible for the commands that insert data in the database instance
//
// The insert package is all about enabling data generation through templates or other commands using the `parameters` package in CLI form
package insert

import (
	"github.com/spf13/cobra"
	"log"
)

var (
	ConnString    string
	table         string
	template      string
	amount        int
	defaultConfig bool
)

// InsertCmd represents the insert command
var InsertCmd = &cobra.Command{
	Use:   "insert",
	Short: "The palette responsible for generating data.",
	Long: `This palette is responsible for generating data,
this can either be generating database table entries with the entry command or 
INSERT and DELETE queries using the queries command
`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func addSubCommands() {
	InsertCmd.AddCommand(entryCmd)
}

func init() {

	InsertCmd.PersistentFlags().StringVarP(&ConnString, "database", "", "", "url to connect to the database with")
	InsertCmd.PersistentFlags().StringVarP(&table, "table", "", "", "name of table in the database")
	InsertCmd.PersistentFlags().StringVarP(&template, "template", "", "", "path to the template file")
	InsertCmd.PersistentFlags().IntVarP(&amount, "amount", "", 1, "amount of items to insert")
	InsertCmd.PersistentFlags().BoolVarP(&defaultConfig, "default", "", false, "flag that determines if the default configuration is used")

	if err := InsertCmd.MarkPersistentFlagRequired("database"); err != nil {
		log.Fatal(err)
	}
	if err := InsertCmd.MarkPersistentFlagRequired("table"); err != nil {
		log.Fatal(err)
	}

	InsertCmd.MarkFlagsOneRequired("template", "default")
	InsertCmd.MarkFlagsMutuallyExclusive("template", "default") // either use a template or use the default

	addSubCommands()
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// InsertCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// InsertCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
