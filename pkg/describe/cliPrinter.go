package describe

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func PrintResourceTable(rootResource Resource) {
	// Create a new table.
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"PARENT", "KIND", "NAME", "NAMESPACE", "STATUS"})

	// Print root resource.
	printResourceAndChildren(table, rootResource, "", "")

	// Render the table.
	table.Render()
}

func printResourceAndChildren(table *tablewriter.Table, resource Resource, parentKind string, parentNamespace string) {
	// Extract information from the resource.
	kind := resource.manifest.GetKind()
	name := resource.manifest.GetName()
	namespace := resource.manifest.GetNamespace()

	// Get status field from the "status" subfield of the object.
	statusField, _, _ := unstructured.NestedFieldCopy(resource.manifest.Object, "status")

	// Extract the "phase" from the "status" subfield if it exists.
	status := ""
	if statusField != nil {
		statusMap, _ := statusField.(map[string]interface{})
		if phase, exists := statusMap["phase"]; exists {
			status = fmt.Sprintf("%v", phase)
		}
	}

	// Determine the parent information for the current resource.
	var parentPrefix string
	if parentKind != "" {
		parentPrefix = fmt.Sprintf("%s", parentKind)
	}

	// Create a new row for the resource.
	tableRow := []string{parentPrefix, kind, name, namespace, status}

	// Add the row to the table.
	table.Append(tableRow)

	// Recursively print children with the updated parent information.
	for _, child := range resource.children {
		printResourceAndChildren(table, child, kind, namespace)
	}
}
