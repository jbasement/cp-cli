/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"path/filepath"

	"github.com/jbasement/cp-cli/pkg/describe"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

var Namespace, Kubeconfig string

// describeCmd represents the describe command
var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe a Claim/ Composite resource and all its childs. cp-cli describe <resource-type> <resource-name>",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// if err := someFunc(); err != nil {
		// 	return err
		// }
		// return nil

		if Kubeconfig == "" {
			Kubeconfig = os.Getenv("KUBECONFIG")
		}
		if Kubeconfig == "" {
			Kubeconfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
		}

		describe.Describe(args, Namespace, Kubeconfig)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(describeCmd)

	describeCmd.Flags().StringVarP(&Namespace, "namespace", "n", "default", "k8s namespace")
	describeCmd.Flags().StringVarP(&Kubeconfig, "kubeconfig", "k", "", "Path to Kubeconfig")

}
