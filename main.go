package main

import (
	"fmt"
	"log"

	"github.com/zufardhiyaulhaq/kubernetes-service-netbox-syncer/client"
	"github.com/zufardhiyaulhaq/kubernetes-service-netbox-syncer/model"
	"github.com/zufardhiyaulhaq/kubernetes-service-netbox-syncer/settings"
)

func main() {
	setting, err := settings.NewSettings()
	if err != nil {
		log.Fatalf("Error loading settings: %v", err)
	}

	fmt.Println("Loaded settings")

	netboxClient, err := client.NewNetboxClient(setting)
	if err != nil {
		log.Fatalf("Error initializing Netbox client: %v", err)
	}
	fmt.Println("Initialized Netbox client")

	kubernetesClient, err := client.NewKubernetesClient(setting)
	if err != nil {
		log.Fatalf("Error initializing Kubernetes client: %v", err)
	}
	fmt.Println("Initialized Kubernetes client")

	// fetch the exisitng prefixes
	existingPrefixes, err := kubernetesClient.CreateOrLoadConfiMap()
	if err != nil {
		log.Fatalf("cannot create or load exisitng configmap: %v", err)
	}

	services, err := kubernetesClient.GetKubernetesService()
	if err != nil {
		log.Fatalf("Error fetching Kubernetes services: %v", err)
	}
	fmt.Printf("Fetched %d Kubernetes services\n", len(services))

	// Build a map of existing External IPs for quick lookup
	existingIPMap := make(map[string]model.Prefix)
	for _, prefix := range existingPrefixes {
		existingIPMap[prefix.ExternalIPs] = prefix
	}

	// Build a map of current service External IPs
	serviceIPMap := make(map[string]model.KubernetesService)
	for _, service := range services {
		serviceIPMap[service.ExternalIPs] = service
	}

	// Find services to create (in Kubernetes but not in Netbox)
	var createdServices []model.KubernetesService
	for _, service := range services {
		if _, exists := existingIPMap[service.ExternalIPs]; !exists {
			createdServices = append(createdServices, service)
		}
	}

	// Create prefixes in Netbox for new services
	for _, service := range createdServices {
		fmt.Printf("Creating prefix for service: %s/%s (%s)\n", service.Namespace, service.Name, service.ExternalIPs)
		prefixes, err := netboxClient.CreatePrefix(service)
		if err != nil {
			log.Printf("Error creating prefix in Netbox for service %s/%s: %v", service.Namespace, service.Name, err)
		} else {
			fmt.Printf("Created prefix in Netbox for service %s/%s\n", service.Namespace, service.Name)
			// Add newly created prefixes to existing list
			existingPrefixes = append(existingPrefixes, prefixes...)
			for _, prefix := range prefixes {
				existingIPMap[prefix.ExternalIPs] = prefix
			}
		}
	}

	// Find prefixes to delete (in Netbox but not in Kubernetes)
	var deletedPrefixes []model.Prefix
	for _, existingPrefix := range existingPrefixes {
		if _, exists := serviceIPMap[existingPrefix.ExternalIPs]; !exists {
			deletedPrefixes = append(deletedPrefixes, existingPrefix)
		}
	}

	// Delete stale prefixes from Netbox
	deletedPrefixIDs := make(map[int32]bool)
	for _, deletedPrefix := range deletedPrefixes {
		err := netboxClient.DeletePrefix(deletedPrefix.PrefixID)
		if err != nil {
			log.Printf("Error deleting prefix %d from Netbox: %v", deletedPrefix.PrefixID, err)
		} else {
			fmt.Printf("Deleted prefix %d from Netbox\n", deletedPrefix.PrefixID)
			deletedPrefixIDs[deletedPrefix.PrefixID] = true
		}
	}

	// Build updated prefixes list (existing prefixes minus deleted ones)
	var updatedPrefixes []model.Prefix
	for _, prefix := range existingPrefixes {
		if !deletedPrefixIDs[prefix.PrefixID] {
			updatedPrefixes = append(updatedPrefixes, prefix)
		}
	}

	fmt.Printf("Updating ConfigMap with %d prefixes\n", len(updatedPrefixes))

	// update the latest prefixes to configmap
	err = kubernetesClient.SavePrefixToConfigMap(updatedPrefixes)
	if err != nil {
		log.Printf("Error saving prefix to ConfigMap: %v", err)
	}
}
