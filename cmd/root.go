/*
Copyright © 2025 Keith Compere <KeithCompere150@gmail.com>
*/
package cmd

import (
	"github.com/Keith1039/Capstone_Test/cmd/generate"
	"github.com/Keith1039/Capstone_Test/cmd/validate"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dbvg",
	Short: "A CLI designed to simplify validating database schemas and generating SQL queries",
	Long: `dbvg is a CLI designed to simplify validating databases and generating SQL queries.
	The validation provided is getting rid of cycles in database schemas. The CLI also provides
	tools to generate table entries which will maintain dependencies.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func addSubCommandPalettes() {
	rootCmd.AddCommand(validate.ValidateCmd)
	rootCmd.AddCommand(generate.GenerateCmd)
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.Capstone_Test.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	addSubCommandPalettes()
}
