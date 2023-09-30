package resource

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
)

func PrintResourceTable(rootResource Resource, fields []string) {
	// Create a new table.
	table := tablewriter.NewWriter(os.Stdout)

	table.SetHeader(fields)
	table.SetColWidth(50)
	printResourceAndChildren(table, fields, rootResource, "")
	table.Render()
}

func printResourceAndChildren(table *tablewriter.Table, fields []string, r Resource, parentKind string) {
	var tableRow = make([]string, len(fields))

	// Using this for loop and if statement approach ensures keeping the same output order as the fields argument was passed
	for i, field := range fields {
		if field == "parent" {
			var parentPrefix string
			if parentKind != "" {
				parentPrefix = fmt.Sprintf("%s", parentKind)
			}
			tableRow[i] = parentPrefix
		}
		if field == "name" {
			tableRow[i] = r.GetName()
		}
		if field == "kind" {
			tableRow[i] = r.GetKind()
		}
		if field == "apiversion" {
			tableRow[i] = r.GetApiVersion()
		}
		if field == "synced" {
			tableRow[i] = r.GetConditionStatus("Synced")
		}
		if field == "ready" {
			tableRow[i] = r.GetConditionStatus("Ready")
		}
		if field == "message" {
			tableRow[i] = r.GetConditionMessage()
		}
		if field == "event" {
			tableRow[i] = r.GetEvent()
		}
	}

	// Add the row to the table.
	table.Append(tableRow)

	// Recursively print children with the updated parent information.
	for _, child := range r.children {
		printResourceAndChildren(table, fields, child, r.GetKind())
	}
}
