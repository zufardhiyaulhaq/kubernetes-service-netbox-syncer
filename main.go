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

	fmt.Printf("Loaded settings: %+v\n", setting)

	netboxClient, err := client.NewNetboxClient(setting)
	if err != nil {
		log.Fatalf("Error initializing Netbox client: %v", err)
	}
	fmt.Printf("Initialized Netbox client: %+v\n", netboxClient)

	kubernetesClient, err := client.NewKubernetesClient(setting)
	if err != nil {
		log.Fatalf("Error initializing Kubernetes client: %v", err)
	}
	fmt.Printf("Initialized Kubernetes client: %+v\n", kubernetesClient)

	services, err := kubernetesClient.GetKubernetesService()
	if err != nil {
		log.Fatalf("Error fetching Kubernetes services: %v", err)
	}
	fmt.Printf("Fetched %d Kubernetes services\n", len(services))

	for _, svc := range services {
		fmt.Printf("Service: %s, %s, %s", svc.ExternalIPs, svc.Name, svc.Namespace)
	}

	// netboxClient.Client().IpamAPI.IpamPrefixesCreate(context.Background()).WritablePrefixRequest(netbox.WritablePrefixRequest{
	// 	Prefix: "",
	// })
}
