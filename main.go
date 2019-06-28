package main

import (
	"flag"
	"os"
	"strconv"

	//"strings"
	//"time"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Parse arguments
	minStr := flag.String("min", "1", "Minimum number of free agents to keep alive. Minimum of 1.")

	flag.Parse()

	min, err := strconv.Atoi(*minStr)

	if err != nil {
		panic("Error converting --min argument to int: " + err.Error())
	} else if min < 1 {
		panic("Error - --min argument cannot be less than 1.")
	}

	// Initialize Kubernetes client
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", home()+"/.kube")
		if err != nil {
			panic("Error initializing Kubernetes config: " + err.Error())
		}
	}

	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil || k8sClient == nil {
		panic("Error initializing Kubernetes config: " + err.Error())
	}
}

func home() string {
	home := os.Getenv("HOME")
	if home != "" {
		return home
	}
	return os.Getenv("USERPROFILE") // windows
}
