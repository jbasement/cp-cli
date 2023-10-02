package resource

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Resource struct {
	manifest *unstructured.Unstructured
	children []Resource
	event    string
}

func (r Resource) GetKind() string {
	return r.manifest.GetKind()
}

func (r Resource) GetName() string {
	return r.manifest.GetName()
}

func (r Resource) GetNamespace() string {
	return r.manifest.GetNamespace()
}

func (r Resource) GetApiVersion() string {
	return r.manifest.GetAPIVersion()
}

func (r Resource) GetConditionStatus(conditionKey string) string {
	conditions, _, _ := unstructured.NestedSlice(r.manifest.Object, "status", "conditions")
	for _, condition := range conditions {
		conditionMap, _ := condition.(map[string]interface{})
		conditionType, _ := conditionMap["type"].(string)
		conditionStatus, _ := conditionMap["status"].(string)

		if conditionType == conditionKey {
			return conditionStatus
		}
	}
	return ""
}

func (r Resource) GetConditionMessage() string {
	conditions, _, _ := unstructured.NestedSlice(r.manifest.Object, "status", "conditions")

	for _, item := range conditions {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if message, exists := itemMap["message"]; exists {
				if messageStr, ok := message.(string); ok {
					return messageStr
				}
			}
		}
	}

	return ""
}

func (r Resource) GetEvent() string {
	return r.event
}

func (r Resource) GotChildren() bool {
	if len(r.children) > 0 {
		return true
	} else {
		return false
	}
}
