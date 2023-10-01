/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jbasement/cp-cli/pkg/resource"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
	"k8s.io/client-go/util/homedir"
)

var Namespace, Kubeconfig, Output, GraphPath string
var Fields, AllowedFields, AllowedOutput []string

// describeCmd represents the describe command
var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe a Claim/ Composite resource and all its children.",
	Long: `Describe a Claim/ Composite resource and all its children.

Command Usage:
	cp-cli describe TYPE[.GROUP] NAME [-n| --namespace NAMESPACE]

Example: 
	cp-cli describe objectstorage my-object-storage 
	cp-cli describe xobjectstorage.my-fqdn.cloud/v1alpha1 my-object-storage -n my-namespace -o graph -f name,kind,ready,synced -p ./myGraph.png

	`,
	Args:         cobra.ExactArgs(2),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if fields are valid
		for _, field := range Fields {
			if !slices.Contains(AllowedFields, field) {
				return fmt.Errorf("Invalid field set: %s\nField has to be one of: %s", field, AllowedFields)
			}
		}

		// Check if output format is valid
		if !slices.Contains(AllowedOutput, Output) {
			return fmt.Errorf("Invalid ouput set: %s\nOutput has to be one of: %s", Output, AllowedOutput)
		}

		if Kubeconfig == "" {
			Kubeconfig = os.Getenv("KUBECONFIG")
		}
		if Kubeconfig == "" {
			Kubeconfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
		}

		// Get a resource object. Contains k8s resource and all its children, also as resource.
		root, err := resource.GetResource(args, Namespace, Kubeconfig)
		if err != nil {
			return fmt.Errorf("Error getting resource -> %w", err)
		}

		// Print out resource
		switch Output {
		case "cli":
			if err := resource.PrintResourceTable(*root, Fields); err != nil {
				return fmt.Errorf("Error printing CLI table: %w\n", err)
			}
		case "graph":
			printer := resource.NewGraphPrinter()
			if err := printer.Print(*root, Fields, GraphPath); err != nil {
				return fmt.Errorf("Error printing graph: %w\n", err)
			}
		}

		return nil
	},
}

func init() {
	AllowedFields = []string{"parent", "name", "kind", "namespace", "apiversion", "synced", "ready", "message", "event"}
	fieldFlagDescription := fmt.Sprintf("Comma-separated list of fields. Available fields are %s", AllowedFields)
	AllowedOutput = []string{"cli", "graph"}
	outputFlagDescription := fmt.Sprintf("Output format of resource. Must be one of %s", AllowedOutput)

	rootCmd.AddCommand(describeCmd)

	describeCmd.Flags().StringVarP(&Namespace, "namespace", "n", "default", "k8s namespace")
	describeCmd.Flags().StringVarP(&Kubeconfig, "kubeconfig", "k", "", "Path to Kubeconfig")
	describeCmd.Flags().StringVarP(&Output, "output", "o", "cli", outputFlagDescription)
	describeCmd.Flags().StringSliceVarP(&Fields, "fields", "f", []string{"parent", "kind", "name", "synced", "ready"}, fieldFlagDescription)
	describeCmd.Flags().StringVarP(&GraphPath, "path", "p", "./graph.png", "Set output path and filename for graph PNG. Must be absolute path and filename must end on '.png'")
}
