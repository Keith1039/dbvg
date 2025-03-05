// Package generate is responsible for the commands that generate data in the database instance
//
// The generate package is all about enabling data generation through templates or other commands using the `parameters` package in CLI form
package generate

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

// GenerateCmd represents the generate command
var GenerateCmd = &cobra.Command{
	Use:   "generate",
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
	GenerateCmd.AddCommand(entryCmd)
	GenerateCmd.AddCommand(queriesCmd)
}

func init() {

	GenerateCmd.PersistentFlags().StringVarP(&ConnString, "database", "", "", "url to connect to the database with")
	GenerateCmd.PersistentFlags().StringVarP(&table, "table", "", "", "name of sql table in the database")
	GenerateCmd.PersistentFlags().StringVarP(&template, "template", "", "", "path to the template file")
	GenerateCmd.PersistentFlags().IntVarP(&amount, "amount", "", 1, "amount of items to generate")
	GenerateCmd.PersistentFlags().BoolVarP(&defaultConfig, "default", "", false, "flag that determines if the default configuration is used")

	if err := GenerateCmd.MarkPersistentFlagRequired("database"); err != nil {
		log.Fatal(err)
	}
	if err := GenerateCmd.MarkPersistentFlagRequired("table"); err != nil {
		log.Fatal(err)
	}

	GenerateCmd.MarkFlagsOneRequired("template", "default")
	GenerateCmd.MarkFlagsMutuallyExclusive("template", "default") // either use a template or use the default

	addSubCommands()
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
