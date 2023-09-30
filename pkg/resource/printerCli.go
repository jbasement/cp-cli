package resource

import (
	"flag"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
)

var (
	showAPIVersion bool
	showReady      bool
	showSynced     bool
	showNamespace  bool
)

func init() {
	// Define flags for additional fields.
	flag.BoolVar(&showAPIVersion, "show-api-version", false, "Show API Version")
	flag.BoolVar(&showReady, "show-ready", false, "Show Ready Status")
	flag.BoolVar(&showSynced, "show-synced", false, "Show Synced Status")
	flag.Parse()
}

func PrintResourceTable(rootResource Resource, flags ...bool) {
	// Create a new table.
	table := tablewriter.NewWriter(os.Stdout)

	showNamespace, showAPIVersion = true, true

	// Define table headers based on selected fields.
	headers := []string{"PARENT", "KIND", "NAME"}
	if showAPIVersion {
		headers = append(headers, "API VERSION")
	}
	if showNamespace {
		headers = append(headers, "NAMESPACE")
	}
	if showReady {
		headers = append(headers, "READY")
	}
	if showSynced {
		headers = append(headers, "SYNCED")
	}

	table.SetHeader(headers)

	printResourceAndChildren(table, rootResource, "")
	table.Render()
}

func printResourceAndChildren(table *tablewriter.Table, r Resource, parentKind string) {
	// Extract information from the resource.
	kind := r.GetKind()
	name := r.GetName()

	// Determine the parent information for the current resource.
	var parentPrefix string
	if parentKind != "" {
		parentPrefix = fmt.Sprintf("%s", parentKind)
	}

	// Create a new row for the resource.
	var tableRow []string
	tableRow = append(tableRow, parentPrefix, kind, name)
	if showAPIVersion {
		tableRow = append(tableRow, r.GetApiVersion())
	}
	if showNamespace {
		tableRow = append(tableRow, r.GetNamespace())
	}

	// Add additional fields based on flags.
	if showReady {
		// You need to implement logic to determine if the resource is ready.
		// You can access the "status" field and check the readiness condition.
		// Example: ready := determineReadyStatus(resource.manifest)
		ready := "true" // Replace with actual logic
		tableRow = append(tableRow, ready)
	}

	if showSynced {
		synced := "true" // Replace with actual logic
		tableRow = append(tableRow, synced)
	}

	// Add the row to the table.
	table.Append(tableRow)

	// Recursively print children with the updated parent information.
	for _, child := range r.children {
		printResourceAndChildren(table, child, kind)
	}
}
