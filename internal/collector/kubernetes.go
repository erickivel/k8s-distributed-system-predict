package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesCollector struct {
	kubernetes_client kubernetes.Clientset
}

func (k *KubernetesCollector) Start(kube_config_file_path string) {
	config, err := clientcmd.BuildConfigFromFlags(
		"",
		kube_config_file_path,
	)

	if err != nil {
		fmt.Println("Error when getting k8s configuration:", err.Error())
		return
	}

	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		fmt.Println("Error creating Kubernetes client:", err.Error())
		return
	}

	k.kubernetes_client = *clientset
}

func (k *KubernetesCollector) PrintAllPods(namespace string) {
	pods, err := k.kubernetes_client.CoreV1().Pods(namespace).List(context.Background(), v1.ListOptions{})

	if err != nil {
		fmt.Println("Error listing pods:", err.Error())
		return
	}
	fmt.Printf("Found %d pods in default namespace\n", len(pods.Items))

	jsonData, err := json.MarshalIndent(pods.Items[0], "", "  ")
	if err != nil {
		fmt.Println("Error marshalling pod:", err.Error())
		return
	}

	fmt.Println("Found pods:\n", string(jsonData))
}
