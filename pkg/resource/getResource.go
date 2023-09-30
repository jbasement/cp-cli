package resource

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

type NodeState string

type KubeClient struct {
	dclient   *dynamic.DynamicClient
	clientset *kubernetes.Clientset
	rmapper   meta.RESTMapper
	dc        *discovery.DiscoveryClient
}

func GetResource(args []string, namespace string, kubeconfig string) Resource {
	kubeClient, err := newKubeClient(kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	root := Resource{
		manifest: &unstructured.Unstructured{Object: map[string]interface{}{}},
	}
	root.manifest, err = kubeClient.getManifest(args[1], args[0], "", namespace)
	if err != nil {
		panic(err.Error())
	}

	root = kubeClient.getChildren(root)

	return root
}

func (kc *KubeClient) getManifest(resourceName string, resourceKind string, apiVersion string, namespace string) (*unstructured.Unstructured, error) {
	gr := schema.ParseGroupResource(resourceKind)
	manifest := &unstructured.Unstructured{
		Object: map[string]interface{}{},
	}

	manifest.SetName(resourceName)
	manifest.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   gr.Group,
		Version: apiVersion,
		Kind:    gr.Resource,
	})

	isNamespaced, err := kc.IsResourceNamespaced(gr.Resource, apiVersion)
	if err != nil {
		return nil, err
	}

	if isNamespaced {
		manifest.SetNamespace(namespace)
	}

	gvr, err := kc.rmapper.ResourceFor(schema.GroupVersionResource{
		Group:    manifest.GroupVersionKind().Group,
		Version:  manifest.GroupVersionKind().Version,
		Resource: manifest.GetKind(),
	})
	if err != nil {
		return nil, err
	}

	result, err := kc.dclient.Resource(gvr).Namespace(manifest.GetNamespace()).Get(context.TODO(), manifest.GetName(), metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (kc *KubeClient) getChildren(resource Resource) Resource {
	if resourceRefMap, found, err := getStringMapFromNestedField(*resource.manifest, "spec", "resourceRef"); found && err == nil {
		resource = kc.setChildren(resourceRefMap, resource)
	} else if resourceRefs, found, err := getSliceOfMapsFromNestedField(*resource.manifest, "spec", "resourceRefs"); found && err == nil {
		for _, resourceRefMap := range resourceRefs {
			resource = kc.setChildren(resourceRefMap, resource)
		}
	}

	return resource
}

func (kc *KubeClient) setChildren(resourceRefMap map[string]string, resource Resource) Resource {
	// Get info about child
	name := resourceRefMap["name"]
	kind := resourceRefMap["kind"]
	apiVersion := resourceRefMap["apiVersion"]

	// Get manifest
	u, err := kc.getManifest(name, kind, apiVersion, "default")
	if err != nil {
		panic(err.Error())
	}

	// Get event
	event := kc.getEvent(name, kind, apiVersion, "default")

	// Set child
	child := Resource{
		manifest: u,
		event:    event,
	}
	// Get children of children
	child = kc.getChildren(child)
	resource.children = append(resource.children, child)

	return resource
}

func (kc *KubeClient) IsResourceNamespaced(resourceKind string, apiVersion string) (bool, error) {
	// This function currently does NOT consider different versions of a resource kind. That may cause issues as the scope of a resource might chance depending on the version.

	// Retrieve the API resource list
	apiResourceLists, err := kc.dc.ServerPreferredResources()
	if err != nil {
		return false, err
	}

	// Trim version if set
	apiVersion = strings.Split(apiVersion, "/")[0]

	for _, apiResourceList := range apiResourceLists {
		for _, apiResource := range apiResourceList.APIResources {
			if apiResource.Group == apiVersion || apiVersion == "" {
				resourceKind = strings.ToLower(resourceKind)
				if apiResource.Name == resourceKind || apiResource.SingularName == resourceKind {
					return apiResource.Namespaced, nil
				}
			}

		}
	}
	// If the resource is not found, return an error or false depending on your needs
	return false, fmt.Errorf("resource not found")
}

func (kc *KubeClient) getEvent(resourceName string, resourceKind string, apiVersion string, namespace string) string {
	// List events for the resource.
	eventList, err := kc.clientset.CoreV1().Events(namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=%s,involvedObject.apiVersion=%s", resourceName, resourceKind, apiVersion),
	})

	if err != nil {
		log.Fatalf("Error listing events: %v", err)
	}

	// Check if there are any events.
	if len(eventList.Items) == 0 {
		return ""
	}

	// Get the latest event.
	latestEvent := eventList.Items[0]
	return latestEvent.Message
}
func newKubeClient(kubeconfig string) (*KubeClient, error) {
	// Initialize a Kubernetes client.
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	// Use to get custom resources
	dclient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// Use to discover API resources
	dc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	// Use to get events
	clientset, _ := kubernetes.NewForConfig(config)

	discoveryCacheDir := filepath.Join("./.kube", "cache", "discovery")
	httpCacheDir := filepath.Join("./.kube", "http-cache")
	discoveryClient, err := disk.NewCachedDiscoveryClientForConfig(
		config,
		discoveryCacheDir,
		httpCacheDir,
		10*time.Minute)
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	rMapper := restmapper.NewShortcutExpander(mapper, discoveryClient)

	return &KubeClient{
		dclient:   dclient,
		clientset: clientset,
		rmapper:   rMapper,
		dc:        dc,
	}, nil
}

func getStringMapFromNestedField(obj unstructured.Unstructured, fields ...string) (map[string]string, bool, error) {
	nestedField, found, err := unstructured.NestedStringMap(obj.Object, fields...)
	if !found || err != nil {
		return nil, false, err
	}

	result := make(map[string]string)
	for key, value := range nestedField {
		result[key] = value
	}

	return result, true, nil
}

func getSliceOfMapsFromNestedField(obj unstructured.Unstructured, fields ...string) ([]map[string]string, bool, error) {
	nestedField, found, err := unstructured.NestedFieldNoCopy(obj.Object, fields...)
	if !found || err != nil {
		return nil, false, err
	}

	var result []map[string]string
	if slice, ok := nestedField.([]interface{}); ok {
		for _, item := range slice {
			if m, ok := item.(map[string]interface{}); ok {
				stringMap := make(map[string]string)
				for key, value := range m {
					if str, ok := value.(string); ok {
						stringMap[key] = str
					}
				}
				result = append(result, stringMap)
			}
		}
	}

	return result, true, nil
}
