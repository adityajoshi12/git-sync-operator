package controllers

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func KubeSync() {
	// Set the path to the Kubernetes configuration file
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// Load the Kubernetes configuration from file
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// Create a Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Create a dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// Read the YAML file containing the Kubernetes manifest
	manifestPath := "./manifest.yaml"
	manifestBytes, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		panic(err.Error())
	}

	// Convert the YAML file to an Unstructured object
	manifest := &unstructured.Unstructured{}
	err = runtime.DecodeInto(schema.GroupVersionKind{}, manifestBytes, manifest)
	if err != nil {
		panic(err.Error())
	}

	// Get the GVK of the object in the manifest
	gvk := manifest.GroupVersionKind()

	// Check if the object already exists in the cluster
	namespaced := true
	namespace := "default"
	_, err = dynamicClient.Resource(gvk).Namespace(namespace).Get(context.Background(), manifest.GetName(), metav1.GetOptions{})
	if err == nil {
		// Object already exists, update it
		_, err = dynamicClient.Resource(gvk).Namespace(namespace).Update(context.Background(), manifest, metav1.UpdateOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("Object updated successfully!")
	} else if errors.IsNotFound(err) {
		// Object does not exist, create it
		_, err = dynamicClient.Resource(gvk).Namespace(namespace).Create(context.Background(), manifest, metav1.CreateOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("Object created successfully!")
	} else {
		panic(err.Error())
	}
}
