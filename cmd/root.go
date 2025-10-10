package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ollie",
	Short: "A CLI helper toolset for Ollama",
	Long: `Ollie is a command-line interface tool that provides utility functions
for working with Ollama. It offers various commands to make your Ollama
experience more convenient and efficient.`,
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
