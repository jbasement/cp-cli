package describe

import (
	"context"
	"fmt"
	"os"
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
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type KubeClient struct {
	client  *dynamic.DynamicClient
	rmapper meta.RESTMapper
	dc      *discovery.DiscoveryClient
}

type Resource struct {
	manifest *unstructured.Unstructured
	children []Resource
}

func Describe(args []string) {
	namespace := "default"

	kubeClient, err := newKubeClient()
	if err != nil {
		panic(err.Error())
	}

	root := Resource{
		manifest: &unstructured.Unstructured{Object: map[string]interface{}{}},
	}
	root.manifest, err = kubeClient.getManifest(args[2], args[0], args[1], namespace)
	if err != nil {
		panic(err.Error())
	}

	root = kubeClient.getChildren(root)
	printResourceHierarchy(root, 5)
}

func printResourceHierarchy(resource Resource, indentLevel int) {
	// Print the resource's manifest with proper indentation
	indent := ""
	for i := 0; i < indentLevel; i++ {
		indent += "  " // Use two spaces for each level of indentation
	}
	fmt.Printf("%sResource Type: %s\n", indent, resource.manifest.GetKind())

	// Recursively print child resources
	for _, child := range resource.children {
		printResourceHierarchy(child, indentLevel+1)
	}
}

func (kc *KubeClient) getManifest(resourceName string, resourceGroup string, apiVersion string, namespace string) (*unstructured.Unstructured, error) {
	gr := schema.ParseGroupResource(resourceGroup)
	fmt.Printf("\nGR ist: %s", gr)

	manifest := &unstructured.Unstructured{
		Object: map[string]interface{}{},
	}
	manifest.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   gr.Group,
		Version: apiVersion,
		Kind:    gr.Resource,
	})
	manifest.SetName(resourceName)
	fmt.Printf("\nManifest ist: %s", manifest)

	isNamespaced, err := kc.IsResourceNamespaced(gr.Resource)
	if err != nil {
		return nil, err
	}

	if isNamespaced {
		manifest.SetNamespace(namespace)
	}

	fmt.Printf("\nGroup ist: %s", manifest.GroupVersionKind().Group)
	fmt.Printf("\nVersion ist: %s", manifest.GroupVersionKind().Version)
	fmt.Printf("\nKind ist: %s", manifest.GetKind())

	gvr, err := kc.rmapper.ResourceFor(schema.GroupVersionResource{
		Group:    manifest.GroupVersionKind().Group,
		Version:  manifest.GroupVersionKind().Version,
		Resource: manifest.GetKind(),
	})
	fmt.Printf("\nGVR ist: %s", gvr)
	if err != nil {
		return nil, err
	}

	result, err := kc.client.Resource(gvr).Namespace(manifest.GetNamespace()).Get(context.TODO(), manifest.GetName(), metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (kc *KubeClient) getChildren(resource Resource) Resource {

	if resourceRefMap, found, err := getStringMapFromNestedField(*resource.manifest, "spec", "resourceRef"); found && err == nil {
		// Get info about child
		name := resourceRefMap["name"]
		kind := resourceRefMap["kind"]
		apiVersion := resourceRefMap["apiVersion"]
		u, err := kc.getManifest(name, kind, apiVersion, "default")
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("\nIm object %s und name %s", kind, name)
		// Set child
		child := Resource{
			manifest: u,
		}
		resource.children = append(resource.children, child)
		// Get children of children
		child = kc.getChildren(child)

	} else if resourceRefs, found, err := getSliceOfMapsFromNestedField(*resource.manifest, "spec", "resourceRefs"); found && err == nil {
		for _, resourceRefMap := range resourceRefs {
			// Get info about child
			name := resourceRefMap["name"]
			kind := resourceRefMap["kind"]
			apiVersion := resourceRefMap["apiVersion"]
			u, err := kc.getManifest(name, kind, apiVersion, "default")
			if err != nil {
				panic(err.Error())
			}
			fmt.Printf("\nIm object %s und name %s", kind, name)
			// Set child
			child := Resource{
				manifest: u,
			}
			resource.children = append(resource.children, child)
			// Get children of children
			child = kc.getChildren(child)
		}
	}

	return resource
}

func (kc *KubeClient) IsResourceNamespaced(resourceKind string) (bool, error) {
	// This function currently does NOT consider different versions of a resource kind. That may cause issues as the scope of a resource might chance depending on the version.

	// Retrieve the API resource list
	apiResourceLists, err := kc.dc.ServerPreferredResources()
	if err != nil {
		return false, err
	}

	for _, apiResourceList := range apiResourceLists {
		for _, apiResource := range apiResourceList.APIResources {
			resourceKind = strings.ToLower(resourceKind)
			if apiResource.Name == resourceKind || apiResource.SingularName == resourceKind {
				return apiResource.Namespaced, nil
			}
		}
	}

	// If the resource is not found, return an error or false depending on your needs
	return false, fmt.Errorf("resource not found")
}

func newKubeClient() (*KubeClient, error) {
	var kubeconfig string

	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}
	if kubeconfig == "" {
		kubeconfig = filepath.Join(homedir.HomeDir(), ".kube", "config")
	}

	// Initialize a Kubernetes client.
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// Create a discovery client using the provided config
	dc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

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
		client:  client,
		rmapper: rMapper,
		dc:      dc,
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
