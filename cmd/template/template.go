/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package template

import (
	"log"

	"github.com/spf13/cobra"
)

var (
	ConnString string
	table      string
)

// TemplateCmd represents the template command
var TemplateCmd = &cobra.Command{
	Use:   "template",
	Short: "The palette responsible for creating and updating templates.",
	Long: `This palette is responsible for creating and updating template files.
`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func addSubCommands() {
	TemplateCmd.AddCommand(createCmd)
	TemplateCmd.AddCommand(updateCmd)
}

func init() {
	TemplateCmd.PersistentFlags().StringVarP(&ConnString, "database", "", "", "url to connect to the database with")
	TemplateCmd.PersistentFlags().StringVarP(&table, "table", "", "", "the name of the sql table that the template is based off of")

	if err := TemplateCmd.MarkPersistentFlagRequired("database"); err != nil {
		log.Fatal(err)
	}
	if err := TemplateCmd.MarkPersistentFlagRequired("table"); err != nil {
		log.Fatal(err)
	}
	addSubCommands()

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
