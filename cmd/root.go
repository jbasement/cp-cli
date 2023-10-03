package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var namepace, kubeconfig, output, graphPath, fieldFlagDescription string
var fields, allowedFields, allowedOutput []string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cp-cli",
	Short: "Crossplane CLI",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	allowedFields = []string{"parent", "name", "kind", "namespace", "apiversion", "synced", "ready", "message", "event"}
	fieldFlagDescription = fmt.Sprintf("Comma-separated list of fields. Available fields are %s", allowedFields)
}
