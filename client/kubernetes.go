// Package client provides Kubernetes and Netbox client implementations for syncing services.
package client

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"slices"

	"github.com/zufardhiyaulhaq/kubernetes-service-netbox-syncer/model"
	"github.com/zufardhiyaulhaq/kubernetes-service-netbox-syncer/settings"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesClient struct {
	k8sClient *kubernetes.Clientset
	Settings  settings.Settings
}

func (c *KubernetesClient) Client() *kubernetes.Clientset {
	return c.k8sClient
}

func (c *KubernetesClient) GetKubernetesService() ([]model.KubernetesService, error) {
	var kubernetesServices []model.KubernetesService

	// Get namespaces to query
	namespaces := c.Settings.KubernetesNamespaceFilter
	if len(namespaces) == 0 {
		// If no namespace filter, get all namespaces
		namespaceList, err := c.k8sClient.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, ns := range namespaceList.Items {
			namespaces = append(namespaces, ns.Name)
		}
	}

	// Query services from each namespace
	for _, namespace := range namespaces {
		services, err := c.k8sClient.CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			log.Printf("failed to list services in namespace %s: %v", namespace, err)
			continue
		}

		for _, svc := range services.Items {
			// Filter by service type
			if !c.matchesTypeFilter(svc.Spec.Type) {
				continue
			}

			// Filter by annotations
			if !c.matchesAnnotationFilter(svc.Annotations) {
				continue
			}

			// Filter by labels
			if !c.matchesLabelFilter(svc.Labels) {
				continue
			}

			// Get external IPs
			externalIP := c.getExternalIP(&svc)
			if externalIP == "" {
				continue
			}

			kubernetesServices = append(kubernetesServices, model.KubernetesService{
				Name:        svc.Name,
				Namespace:   svc.Namespace,
				ExternalIPs: externalIP,
			})
		}
	}

	return kubernetesServices, nil
}

// matchesTypeFilter checks if the service type matches the filter
func (c *KubernetesClient) matchesTypeFilter(serviceType v1.ServiceType) bool {
	// If no filter, accept all
	if len(c.Settings.KubernetesTypeFilter) == 0 {
		return true
	}

	return slices.Contains(c.Settings.KubernetesTypeFilter, string(serviceType))
}

// matchesAnnotationFilter checks if the service annotations match the filter
func (c *KubernetesClient) matchesAnnotationFilter(annotations map[string]string) bool {
	// If no filter, accept all
	if len(c.Settings.KubernetesServiceAnnotationFilter) == 0 {
		return true
	}

	// Service must match all annotation filters
	for _, filterMap := range c.Settings.KubernetesServiceAnnotationFilter {
		for key, value := range filterMap {
			if annotations[key] != value {
				return false
			}
		}
	}
	return true
}

// matchesLabelFilter checks if the service labels match the filter
func (c *KubernetesClient) matchesLabelFilter(labels map[string]string) bool {
	// If no filter, accept all
	if len(c.Settings.KubernetesServiceLabelFilter) == 0 {
		return true
	}

	// Service must match all label filters
	for _, filterMap := range c.Settings.KubernetesServiceLabelFilter {
		for key, value := range filterMap {
			if labels[key] != value {
				return false
			}
		}
	}
	return true
}

// getExternalIP retrieves the external IP from the service
func (c *KubernetesClient) getExternalIP(svc *v1.Service) string {
	// Try LoadBalancer Ingress first
	if len(svc.Status.LoadBalancer.Ingress) > 0 {
		if svc.Status.LoadBalancer.Ingress[0].IP != "" {
			return svc.Status.LoadBalancer.Ingress[0].IP
		}
		if svc.Status.LoadBalancer.Ingress[0].Hostname != "" {
			return svc.Status.LoadBalancer.Ingress[0].Hostname
		}
	}

	// Try ExternalIPs
	if len(svc.Spec.ExternalIPs) > 0 {
		return svc.Spec.ExternalIPs[0]
	}

	return ""
}

func (c *KubernetesClient) CreateOrLoadConfiMap() ([]model.Prefix, error) {
	var prefixes []model.Prefix

	configMapName := c.Settings.KubernetesConfigMapName
	configMapNamespace := c.Settings.KubernetesConfigMapNamepace

	// Try to get existing ConfigMap
	configMap, err := c.k8sClient.CoreV1().ConfigMaps(configMapNamespace).Get(
		context.Background(),
		configMapName,
		metav1.GetOptions{},
	)

	if err != nil {
		if errors.IsNotFound(err) {
			// ConfigMap not found, create it with empty data
			newConfigMap := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      configMapName,
					Namespace: configMapNamespace,
				},
				Data: map[string]string{
					"prefixes.json": "[]",
				},
			}

			_, err = c.k8sClient.CoreV1().ConfigMaps(configMapNamespace).Create(
				context.Background(),
				newConfigMap,
				metav1.CreateOptions{},
			)
			if err != nil {
				return nil, err
			}
			log.Printf("Created new ConfigMap %s/%s with empty prefix list", configMapNamespace, configMapName)
			return prefixes, nil
		}
		return nil, err
	}

	// ConfigMap exists, load the data
	if data, ok := configMap.Data["prefixes.json"]; ok && data != "" {
		err = json.Unmarshal([]byte(data), &prefixes)
		if err != nil {
			log.Printf("Failed to unmarshal prefixes from ConfigMap: %v", err)
			return nil, err
		}
		log.Printf("Loaded %d prefixes from ConfigMap %s/%s", len(prefixes), configMapNamespace, configMapName)
	} else {
		log.Printf("ConfigMap %s/%s exists but has no prefix data", configMapNamespace, configMapName)
	}

	return prefixes, nil
}

func (c *KubernetesClient) SavePrefixToConfigMap(prefixes []model.Prefix) error {
	// Marshal prefixes to JSON
	data, err := json.MarshalIndent(prefixes, "", "  ")
	if err != nil {
		return err
	}

	configMapName := c.Settings.KubernetesConfigMapName
	configMapNamespace := c.Settings.KubernetesConfigMapNamepace

	// Check if ConfigMap exists
	existingConfigMap, err := c.k8sClient.CoreV1().ConfigMaps(configMapNamespace).Get(
		context.Background(),
		configMapName,
		metav1.GetOptions{},
	)

	if err != nil {
		if errors.IsNotFound(err) {
			// Create new ConfigMap
			configMap := &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      configMapName,
					Namespace: configMapNamespace,
				},
				Data: map[string]string{
					"prefixes.json": string(data),
				},
			}

			_, err = c.k8sClient.CoreV1().ConfigMaps(configMapNamespace).Create(
				context.Background(),
				configMap,
				metav1.CreateOptions{},
			)
			if err != nil {
				return err
			}
			log.Printf("Created ConfigMap %s/%s", configMapNamespace, configMapName)
			return nil
		}
		return err
	}

	// Update existing ConfigMap
	existingConfigMap.Data = map[string]string{
		"prefixes.json": string(data),
	}

	_, err = c.k8sClient.CoreV1().ConfigMaps(configMapNamespace).Update(
		context.Background(),
		existingConfigMap,
		metav1.UpdateOptions{},
	)
	if err != nil {
		return err
	}
	log.Printf("Updated ConfigMap %s/%s", configMapNamespace, configMapName)
	return nil
}

func NewKubernetesClient(settings settings.Settings) (*KubernetesClient, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first
	config, err = rest.InClusterConfig()
	if err != nil {
		log.Print("failed to get in-cluster config, trying kubeconfig")

		// Fallback to kubeconfig
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			// Default kubeconfig location
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, err
			}
			kubeconfig = filepath.Join(home, ".kube", "config")
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	conf := KubernetesClient{
		k8sClient: k8sClient,
		Settings:  settings,
	}
	return &conf, nil
}
