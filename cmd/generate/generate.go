/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package generate

import (
	"database/sql"
	"github.com/spf13/cobra"
	"log"
)

var (
	ConnString string
)

// GenerateCmd represents the generate command
var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func addSubCommands() {
	GenerateCmd.AddCommand(templateCmd)
	GenerateCmd.AddCommand(entryCmd)
}

func init() {

	GenerateCmd.PersistentFlags().StringVarP(&ConnString, "database", "", "", "url to connect to the database with")

	if err := GenerateCmd.MarkPersistentFlagRequired("database"); err != nil {
		log.Fatal(err)
	}
	addSubCommands()
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", ConnString)
	if err != nil {
		log.Fatal(err)
	}
	return db, nil
}
