package collector

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesCollector struct {
	kubernetes_client kubernetes.Clientset
}

type KubernetesNamespace struct {
	Name         string    `json:"name"`
	CreationDate time.Time `json:"creation_date"`
}

type KubernetesNode struct {
	Name   string `json:"name"`
	Cpu    int64  `json:"cpu"`    // millicores
	Memory int64  `json:"memory"` // bytes
}

type KubernetesPod struct {
	Uid           string              `json:"uid"`
	Name          string              `json:"name"`
	Namespace     string              `json:"namespace"`
	NodeName      string              `json:"node_name"`
	Labels        map[string]string   `json:"labels"`
	CreationDate  time.Time           `json:"creation_date"`
	Containers    KubernetesContainer `json:"containers"`
	CpuRequest    int64               `json:"cpu_request"`    // millicores
	CpuLimit      int64               `json:"cpu_limit"`      // millicores
	MemoryRequest int64               `json:"memory_request"` // bytes
	MemoryLimit   int64               `json:"memory_limit"`   // bytes
}

type KubernetesHPA struct {
	Name                    string `json:"name"`
	Namespace               string `json:"namespace"`
	ScaleTargetRefKind      string `json:"scale_target_ref_kind"`
	ScaleTargetRefName      string `json:"scale_target_ref_name"`
	MinReplicas             int32  `json:"min_replicas"`
	MaxReplicas             int32  `json:"max_replicas"`
	TargetCPUUtilization    int32  `json:"target_cpu_utilization"`    // percentage (0-100)
	TargetMemoryUtilization int32  `json:"target_memory_utilization"` // percentage (0-100), 0 if not set
	CurrentReplicas         int32  `json:"current_replicas"`
	DesiredReplicas         int32  `json:"desired_replicas"`
}

type KubernetesService struct {
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	Type         string            `json:"type"` // ClusterIP, NodePort, LoadBalancer
	ClusterIP    string            `json:"cluster_ip"`
	Selector     map[string]string `json:"selector"`
	Ports        []ServicePort     `json:"ports"`
	CreationDate time.Time         `json:"creation_date"`
}

type ServicePort struct {
	Name       string `json:"name"`
	Port       int32  `json:"port"`
	TargetPort int32  `json:"target_port"`
	Protocol   string `json:"protocol"`
}

type KubernetesContainer struct{}

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

func (k *KubernetesCollector) GetAllNamespaces() ([]KubernetesNamespace, error) {
	parsedNamespaces := []KubernetesNamespace{}

	namespaces, err := k.kubernetes_client.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	for _, namespace := range namespaces.Items {
		kubernetes_namespace := KubernetesNamespace{
			Name:         namespace.Name,
			CreationDate: namespace.CreationTimestamp.Time,
		}

		parsedNamespaces = append(parsedNamespaces, kubernetes_namespace)
	}

	return parsedNamespaces, nil
}

func (k *KubernetesCollector) GetAllNodes() ([]KubernetesNode, error) {
	var parsedNodes []KubernetesNode

	nodes, err := k.kubernetes_client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	for _, node := range nodes.Items {
		kNode := KubernetesNode{
			Name:   node.Name,
			Memory: node.Status.Capacity.Memory().Value(),   // bytes
			Cpu:    node.Status.Capacity.Cpu().MilliValue(), // millicores
		}

		parsedNodes = append(parsedNodes, kNode)
	}

	return parsedNodes, nil
}

func (k *KubernetesCollector) GetAllPods() ([]KubernetesPod, error) {
	parsedPods := []KubernetesPod{}

	pods, err := k.kubernetes_client.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		kubernetes_pod := KubernetesPod{
			Uid:          string(pod.UID),
			Name:         pod.Name,
			Namespace:    pod.Namespace,
			NodeName:     pod.Spec.NodeName,
			Labels:       pod.Labels,
			CreationDate: pod.CreationTimestamp.Time,
			Containers:   KubernetesContainer{},
		}

		// Extract data from the first container
		if len(pod.Spec.Containers) > 0 {
			container := pod.Spec.Containers[0]
			kubernetes_pod.CpuRequest = container.Resources.Requests.Cpu().MilliValue()  // millicores
			kubernetes_pod.CpuLimit = container.Resources.Limits.Cpu().MilliValue()      // millicores
			kubernetes_pod.MemoryRequest = container.Resources.Requests.Memory().Value() // bytes
			kubernetes_pod.MemoryLimit = container.Resources.Limits.Memory().Value()     // bytes
		}

		parsedPods = append(parsedPods, kubernetes_pod)
	}

	return parsedPods, nil
}

func (k *KubernetesCollector) GetAllHPAs() ([]KubernetesHPA, error) {
	var parsedHPAs []KubernetesHPA

	hpas, err := k.kubernetes_client.AutoscalingV2().HorizontalPodAutoscalers("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, hpa := range hpas.Items {
		kubernetesHPA := KubernetesHPA{
			Name:               hpa.Name,
			Namespace:          hpa.Namespace,
			ScaleTargetRefKind: hpa.Spec.ScaleTargetRef.Kind,
			ScaleTargetRefName: hpa.Spec.ScaleTargetRef.Name,
			MaxReplicas:        hpa.Spec.MaxReplicas,
			CurrentReplicas:    hpa.Status.CurrentReplicas,
			DesiredReplicas:    hpa.Status.DesiredReplicas,
		}

		// MinReplicas is a pointer, default to 1 if not set
		if hpa.Spec.MinReplicas != nil {
			kubernetesHPA.MinReplicas = *hpa.Spec.MinReplicas
		} else {
			kubernetesHPA.MinReplicas = 1
		}

		// Extract CPU and Memory utilization targets from metrics
		for _, metric := range hpa.Spec.Metrics {
			if metric.Resource != nil {
				if metric.Resource.Name == "cpu" && metric.Resource.Target.AverageUtilization != nil {
					kubernetesHPA.TargetCPUUtilization = *metric.Resource.Target.AverageUtilization
				}
				if metric.Resource.Name == "memory" && metric.Resource.Target.AverageUtilization != nil {
					kubernetesHPA.TargetMemoryUtilization = *metric.Resource.Target.AverageUtilization
				}
			}
		}

		parsedHPAs = append(parsedHPAs, kubernetesHPA)
	}

	return parsedHPAs, nil
}

func (k *KubernetesCollector) GetHPAsByNamespace(namespace string) ([]KubernetesHPA, error) {
	var parsedHPAs []KubernetesHPA

	hpas, err := k.kubernetes_client.AutoscalingV2().HorizontalPodAutoscalers(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, hpa := range hpas.Items {
		kubernetesHPA := KubernetesHPA{
			Name:               hpa.Name,
			Namespace:          hpa.Namespace,
			ScaleTargetRefKind: hpa.Spec.ScaleTargetRef.Kind,
			ScaleTargetRefName: hpa.Spec.ScaleTargetRef.Name,
			MaxReplicas:        hpa.Spec.MaxReplicas,
			CurrentReplicas:    hpa.Status.CurrentReplicas,
			DesiredReplicas:    hpa.Status.DesiredReplicas,
		}

		if hpa.Spec.MinReplicas != nil {
			kubernetesHPA.MinReplicas = *hpa.Spec.MinReplicas
		} else {
			kubernetesHPA.MinReplicas = 1
		}

		for _, metric := range hpa.Spec.Metrics {
			if metric.Resource != nil {
				if metric.Resource.Name == "cpu" && metric.Resource.Target.AverageUtilization != nil {
					kubernetesHPA.TargetCPUUtilization = *metric.Resource.Target.AverageUtilization
				}
				if metric.Resource.Name == "memory" && metric.Resource.Target.AverageUtilization != nil {
					kubernetesHPA.TargetMemoryUtilization = *metric.Resource.Target.AverageUtilization
				}
			}
		}

		parsedHPAs = append(parsedHPAs, kubernetesHPA)
	}

	return parsedHPAs, nil
}

func (k *KubernetesCollector) GetAllServices() ([]KubernetesService, error) {
	var parsedServices []KubernetesService

	services, err := k.kubernetes_client.CoreV1().Services("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, svc := range services.Items {
		kubernetesService := KubernetesService{
			Name:         svc.Name,
			Namespace:    svc.Namespace,
			Type:         string(svc.Spec.Type),
			ClusterIP:    svc.Spec.ClusterIP,
			Selector:     svc.Spec.Selector,
			CreationDate: svc.CreationTimestamp.Time,
		}

		for _, port := range svc.Spec.Ports {
			servicePort := ServicePort{
				Name:       port.Name,
				Port:       port.Port,
				TargetPort: port.TargetPort.IntVal,
				Protocol:   string(port.Protocol),
			}
			kubernetesService.Ports = append(kubernetesService.Ports, servicePort)
		}

		parsedServices = append(parsedServices, kubernetesService)
	}

	return parsedServices, nil
}

func (k *KubernetesCollector) GetServicesByNamespace(namespace string) ([]KubernetesService, error) {
	var parsedServices []KubernetesService

	services, err := k.kubernetes_client.CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, svc := range services.Items {
		kubernetesService := KubernetesService{
			Name:         svc.Name,
			Namespace:    svc.Namespace,
			Type:         string(svc.Spec.Type),
			ClusterIP:    svc.Spec.ClusterIP,
			Selector:     svc.Spec.Selector,
			CreationDate: svc.CreationTimestamp.Time,
		}

		for _, port := range svc.Spec.Ports {
			servicePort := ServicePort{
				Name:       port.Name,
				Port:       port.Port,
				TargetPort: port.TargetPort.IntVal,
				Protocol:   string(port.Protocol),
			}
			kubernetesService.Ports = append(kubernetesService.Ports, servicePort)
		}

		parsedServices = append(parsedServices, kubernetesService)
	}

	return parsedServices, nil
}

func (k *KubernetesCollector) GetPodsByNamespace(namespace string) ([]KubernetesPod, error) {
	var parsedPods []KubernetesPod

	pods, err := k.kubernetes_client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		kubernetes_pod := KubernetesPod{
			Uid:          string(pod.UID),
			Name:         pod.Name,
			Namespace:    pod.Namespace,
			NodeName:     pod.Spec.NodeName,
			Labels:       pod.Labels,
			CreationDate: pod.CreationTimestamp.Time,
			Containers:   KubernetesContainer{},
		}

		if len(pod.Spec.Containers) > 0 {
			container := pod.Spec.Containers[0]
			kubernetes_pod.CpuRequest = container.Resources.Requests.Cpu().MilliValue()
			kubernetes_pod.CpuLimit = container.Resources.Limits.Cpu().MilliValue()
			kubernetes_pod.MemoryRequest = container.Resources.Requests.Memory().Value()
			kubernetes_pod.MemoryLimit = container.Resources.Limits.Memory().Value()
		}

		parsedPods = append(parsedPods, kubernetes_pod)
	}

	return parsedPods, nil
}
