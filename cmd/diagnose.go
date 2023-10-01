/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/jbasement/cp-cli/pkg/resource"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

// diagnoseCmd represents the diagnose command
var diagnoseCmd = &cobra.Command{
	Use:          "diagnose",
	Short:        "Diagnose a given resource.",
	Args:         cobra.ExactArgs(2),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if kubeconfig == "" {
			kubeconfig = os.Getenv("KUBECONFIG")
		}
		if kubeconfig == "" {
			kubeconfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
		}

		// Get resource object. Contains k8s resource and all its children, also as resource.
		root, err := resource.GetResource(args, namepace, kubeconfig)
		if err != nil {
			return fmt.Errorf("Error getting resource -> %w", err)
		}

		// Find unhealthy resources
		var unhealthyR resource.Resource
		unhealthyR, err = resource.Diagnose(*root, unhealthyR)
		if err != nil {
			return fmt.Errorf("Couldn't finish diagnose -> %w", err)
		}

		if !reflect.DeepEqual(unhealthyR, resource.Resource{}) {
			// CLI print unhealthy resources
			fmt.Printf("Identified the following resources as potentialy unhealthy.\n")
			if err := resource.PrintResourceTable(unhealthyR, []string{"kind", "apiversion", "name", "synced", "ready", "message", "event"}); err != nil {
				return fmt.Errorf("Error printing CLI table: %w\n", err)
			}
		} else {
			fmt.Printf("Couldn't diagnose any issue with resource %s %s.", root.GetKind(), root.GetName())
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(diagnoseCmd)

	diagnoseCmd.Flags().StringVarP(&namepace, "namespace", "n", "default", "k8s namespace")
	diagnoseCmd.Flags().StringVarP(&kubeconfig, "kubeconfig", "k", "", "Path to Kubeconfig")
}