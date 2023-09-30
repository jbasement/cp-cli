package resource

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

type Resource struct {
	manifest *unstructured.Unstructured
	children []Resource
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
