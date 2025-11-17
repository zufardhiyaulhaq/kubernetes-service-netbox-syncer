package main

import (
	"fmt"
	"log"

	"github.com/zufardhiyaulhaq/kubernetes-service-netbox-syncer/client"
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

	services, err := kubernetesClient.GetKubernetesService()
	if err != nil {
		log.Fatalf("Error fetching Kubernetes services: %v", err)
	}
	fmt.Printf("Fetched %d Kubernetes services\n", len(services))

	for _, svc := range services {
		fmt.Printf("Service: %s, %s, %s\n", svc.ExternalIPs, svc.Name, svc.Namespace)
		err := netboxClient.CreatePrefix(svc)
		if err != nil {
			log.Printf("Error creating prefix in Netbox for service %s/%s: %v", svc.Namespace, svc.Name, err)
		} else {
			fmt.Printf("Created prefix in Netbox for service %s/%s\n", svc.Namespace, svc.Name)
		}
	}
}
