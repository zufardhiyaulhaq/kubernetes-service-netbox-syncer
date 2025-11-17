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

	// loop service to be created in netbox
	var createdServices []model.KubernetesService

	for _, service := range services {
		skipCreated := false

		for _, existingPrefix := range existingPrefixes {
			if service.ExternalIPs == existingPrefix.ExternalIPs {
				skipCreated = true
				continue
			}
		}

		if !skipCreated {
			createdServices = append(createdServices, service)
		}
	}

	// create the servicde in netbox and return prefixes
	for _, service := range createdServices {
		fmt.Printf("Service: %s, %s, %s\n", service.ExternalIPs, service.Name, service.Namespace)
		prefixes, err := netboxClient.CreatePrefix(service)
		if err != nil {
			log.Printf("Error creating prefix in Netbox for service %s/%s: %v", service.Namespace, service.Name, err)
		} else {
			fmt.Printf("Created prefix in Netbox for service %s/%s\n", service.Namespace, service.Name)
		}

		existingPrefixes = append(existingPrefixes, prefixes...)
	}

	// loop into service from kubernetes, if prefix from configmap is not found on kubernetes service based on external IP, we need to delete the prefix from netbox
	var deletedPrefixes []model.Prefix

	for _, existingPrefix := range existingPrefixes {
		deletePrefix := true
		for _, service := range services {
			if existingPrefix.ExternalIPs == service.ExternalIPs {
				deletePrefix = false
				continue
			}
		}

		if deletePrefix {
			deletedPrefixes = append(deletedPrefixes, existingPrefix)
		}
	}

	for _, deletedPrefix := range deletedPrefixes {
		err := netboxClient.DeletePrefix(deletedPrefix.PrefixID)
		if err != nil {
			log.Printf("Error deleting prefix %d from Netbox: %v", deletedPrefix.PrefixID, err)
		} else {
			fmt.Printf("Deleted prefix %d from Netbox\n", deletedPrefix.PrefixID)
		}
	}

	var updatedPrefixes []model.Prefix

	// update the latest prefixes to configmap
	err = kubernetesClient.SavePrefixToConfigMap(updatedPrefixes)
	if err != nil {
		log.Printf("Error saving prefix to ConfigMap: %v", err)
	}
}
