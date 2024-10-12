package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s get <resource>\n", os.Args[0])
		os.Exit(1)
	}
	command := os.Args[1]
	kind := os.Args[2]

	if command != "get" {
		fmt.Println("Unsupported command:", command)
		os.Exit(1)
	}

	home := homedir.HomeDir()
	if home != "" {
		home = filepath.Join(home, ".kube", "config")
	}

	kubeconfig := flag.String("kubeconfig", home, "absolute path to the kubeconfig file")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	discoveryClient := clientSet.Discovery()
	apiGroupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		log.Fatal(err)
	}

	mapper := restmapper.NewDiscoveryRESTMapper(apiGroupResources)
	gvk := schema.FromAPIVersionAndKind("aiops.geektime.com/v1alpha1", kind)

	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		log.Fatal(err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	resourceInterface := dynamicClient.Resource(mapping.Resource).Namespace("default")
	resources, err := resourceInterface.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	for _, resource := range resources.Items {
		fmt.Printf("Name: %s, Namespace: %s, UID: %s\n", resource.GetName(), resource.GetNamespace(), resource.GetUID())
	}
}
