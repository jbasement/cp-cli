package resource

import (
	"reflect"
)

func Diagnose(r Resource, unhealthyR Resource) (Resource, error) {
	// Diagnose self
	if r.GetConditionStatus("Synced") == "False" || r.GetConditionStatus("Ready") == "False" {
		// If first resource is added to unhealthy Resource struct set it as root. Else resource as child.
		if reflect.DeepEqual(unhealthyR, Resource{}) {
			unhealthyR.manifest = r.manifest
			unhealthyR.event = r.event
		} else {
			unhealthyR.children = append(unhealthyR.children, r)
		}
	}
	// Diagnose children
	for _, resource := range r.children {
		unhealthyR, _ = Diagnose(resource, unhealthyR)
	}

	return unhealthyR, nil
}
