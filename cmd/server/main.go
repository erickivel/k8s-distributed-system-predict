package main

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/erickivel/kubeplimsoll/internal/collector"
)

func printKubernetesData(namespaces []collector.KubernetesNamespace, nodes []collector.KubernetesNode, pods []collector.KubernetesPod, services []collector.KubernetesService, hpas []collector.KubernetesHPA) {
	fmt.Println("--------------------------------")
	fmt.Println("Namespaces:")
	fmt.Println("--------------------------------")
	for _, namespace := range namespaces {
		jsonData, err := json.MarshalIndent(namespace, "", "  ")

		if err != nil {
			fmt.Println("Error marshalling namespaces:", err.Error())
			return
		}

		fmt.Println(string(jsonData))
	}
	fmt.Println("--------------------------------")
	fmt.Println("Nodes:")
	fmt.Println("--------------------------------")
	for _, node := range nodes {
		jsonData, err := json.MarshalIndent(node, "", "  ")

		if err != nil {
			fmt.Println("Error marshalling nodes:", err.Error())
			return
		}

		fmt.Println(string(jsonData))
	}

	fmt.Println("--------------------------------")
	fmt.Println("Pods:")
	fmt.Println("--------------------------------")
	for _, pod := range pods {
		jsonData, err := json.MarshalIndent(pod, "", "  ")

		if err != nil {
			fmt.Println("Error marshalling pods:", err.Error())
			return
		}

		fmt.Println(string(jsonData))
	}

	fmt.Println("--------------------------------")
	fmt.Println("Services:")
	fmt.Println("--------------------------------")
	for _, service := range services {
		jsonData, err := json.MarshalIndent(service, "", "  ")

		if err != nil {
			fmt.Println("Error marshalling services:", err.Error())
			return
		}

		fmt.Println(string(jsonData))
	}

	fmt.Println("--------------------------------")
	fmt.Println("HPAs:")
	fmt.Println("--------------------------------")
	for _, hpa := range hpas {
		jsonData, err := json.MarshalIndent(hpa, "", "  ")

		if err != nil {
			fmt.Println("Error marshalling HPAs:", err.Error())
			return
		}

		fmt.Println(string(jsonData))
	}
}

func main() {
	kube_config_file_path := flag.String("kube-config", "", "Kube config file path")

	if kube_config_file_path == nil {
		fmt.Println("Invalid NULL kube_config file path")
		return
	}

	flag.Parse()

	kubernetes_collector := collector.KubernetesCollector{}

	kubernetes_collector.Start(*kube_config_file_path)

	namespaces, err := kubernetes_collector.GetAllNamespaces()

	if err != nil {
		fmt.Println("Error getting all namespaces:", err.Error())
	}

	nodes, err := kubernetes_collector.GetAllNodes()

	if err != nil {
		fmt.Println("Error getting all nodes:", err.Error())
		return
	}

	pods, err := kubernetes_collector.GetAllPods()

	if err != nil {
		fmt.Println("Error getting all pods:", err.Error())
		return
	}

	services, err := kubernetes_collector.GetAllServices()

	if err != nil {
		fmt.Println("Error getting all services:", err.Error())
		return
	}

	hpas, err := kubernetes_collector.GetAllHPAs()

	if err != nil {
		fmt.Println("Error getting all HPAs:", err.Error())
		return
	}

	printKubernetesData(namespaces, nodes, pods, services, hpas)
}
