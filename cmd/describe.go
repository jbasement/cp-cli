/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jbasement/cp-cli/pkg/describe"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

var Namespace, Kubeconfig, Output string

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

		root := describe.Describe(args, Namespace, Kubeconfig)

		switch Output {
		case "cli":
			describe.PrintResourceTable(root)
		case "graph":
			printer := describe.NewGraphPrinter()
			if err := printer.Print([]describe.Resource{root}); err != nil {
				fmt.Printf("Error printing graph: %v\n", err)
			}
		default:
			fmt.Println("Invalid output format. Please use 'cli' or 'graph'.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(describeCmd)

	describeCmd.Flags().StringVarP(&Namespace, "namespace", "n", "default", "k8s namespace")
	describeCmd.Flags().StringVarP(&Kubeconfig, "kubeconfig", "k", "", "Path to Kubeconfig")
	describeCmd.Flags().StringVarP(&Output, "output", "o", "cli", "Output format (cli or graph)")

}
