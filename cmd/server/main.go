package main

import (
	"flag"
	"fmt"
	"github.com/erickivel/kubeplimsoll/internal/collector"
)

func main() {
	kube_config_file_path := flag.String("kube-config", "", "Kube config file path")

	if kube_config_file_path == nil {
		fmt.Println("Invalid NULL kube_config file path")
		return
	}

	flag.Parse()

	kubernetes_collector := collector.KubernetesCollector{}

	kubernetes_collector.Start(*kube_config_file_path)

	kubernetes_collector.PrintAllPods("kube-system")
}
