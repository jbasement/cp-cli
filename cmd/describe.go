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
	"k8s.io/client-go/util/homedir"
)

var Namespace, Kubeconfig, Output string
var Fields []string

// describeCmd represents the describe command
var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe a Claim/ Composite resource and all its children.",
	Long: `Describe a Claim/ Composite resource and all its children.

Command Usage:
	cp-cli describe TYPE[.GROUP] NAME [-n| --namespace NAMESPACE]

Example: 
	cp-cli describe objectstorage my-object-storage 
	cp-cli describe xobjectstorage.my-fqdn.cloud/v1alpha1 my-object-storage -n my-namespace 

	`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if Kubeconfig == "" {
			Kubeconfig = os.Getenv("KUBECONFIG")
		}
		if Kubeconfig == "" {
			Kubeconfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
		}

		// add args and flag parser here

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
			if err := printer.Print(*root, Fields); err != nil {
				return fmt.Errorf("Error printing graph: %w\n", err)
			}
		default:
			return fmt.Errorf("Invalid output format. Please use 'cli' or 'graph'.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(describeCmd)

	describeCmd.Flags().StringVarP(&Namespace, "namespace", "n", "default", "k8s namespace")
	describeCmd.Flags().StringVarP(&Kubeconfig, "kubeconfig", "k", "", "Path to Kubeconfig")
	describeCmd.Flags().StringVarP(&Output, "output", "o", "cli", "Output format (cli or graph)")
	describeCmd.Flags().StringSliceVar(&Fields, "fields", []string{"parent", "kind", "name", "synced", "ready", "message"}, "Comma-separated list of fields")

}
