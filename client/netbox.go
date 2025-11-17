package client

import (
	"context"
	"fmt"

	"github.com/netbox-community/go-netbox/v4"
	"github.com/zufardhiyaulhaq/kubernetes-service-netbox-syncer/model"
	"github.com/zufardhiyaulhaq/kubernetes-service-netbox-syncer/settings"
	"github.com/zufardhiyaulhaq/kubernetes-service-netbox-syncer/utils"
)

type NetboxClient struct {
	netboxClient *netbox.APIClient
	settings     settings.Settings
}

func (c *NetboxClient) Client() *netbox.APIClient {
	return c.netboxClient
}

func (c *NetboxClient) CreatePrefix(service model.KubernetesService) ([]model.Prefix, error) {
	customFields := make(map[string]interface{})
	for _, field := range c.settings.NetboxCustomField {
		for k, v := range field {
			customFields[k] = v
		}
	}

	prefixes := []model.Prefix{}
	markUtilized := true
	isPool := false

	if utils.CheckIP(service.ExternalIPs) {
		description := fmt.Sprintf("%s-%s-%s-%s", service.ExternalIPs, service.Name, service.Namespace, c.settings.KubernetesCluster)

		prefix, _, err := c.netboxClient.IpamAPI.IpamPrefixesCreate(context.Background()).WritablePrefixRequest(netbox.WritablePrefixRequest{
			Prefix:      service.ExternalIPs + "/32",
			Description: &description,

			Status:       netbox.PATCHEDWRITABLEPREFIXREQUESTSTATUS_ACTIVE.Ptr(),
			IsPool:       &isPool,
			MarkUtilized: &markUtilized,
			CustomFields: customFields,
		}).Execute()

		if err != nil {
			return []model.Prefix{}, fmt.Errorf("failed to create prefix in Netbox: %v", err)
		}

		prefixes = append(prefixes, model.Prefix{
			PrefixID:    prefix.Id,
			Prefix:      service.ExternalIPs + "/32",
			ExternalIPs: service.ExternalIPs,
			ServiceName: service.Name,
			Namespace:   service.Namespace,
		})
	}

	if utils.CheckDNS(service.ExternalIPs) {
		IPs, err := utils.GetIPFromDNS(service.ExternalIPs)
		if err != nil {
			return []model.Prefix{}, fmt.Errorf("failed to resolve DNS %s: %v", service.ExternalIPs, err)
		}

		for _, ip := range IPs {
			description := fmt.Sprintf("%s-%s-%s-%s-%s", ip, service.ExternalIPs, service.Name, service.Namespace, c.settings.KubernetesCluster)

			prefix, _, err := c.netboxClient.IpamAPI.IpamPrefixesCreate(context.Background()).WritablePrefixRequest(netbox.WritablePrefixRequest{
				Prefix:      ip + "/32",
				Description: &description,

				Status:       netbox.PATCHEDWRITABLEPREFIXREQUESTSTATUS_ACTIVE.Ptr(),
				IsPool:       &isPool,
				MarkUtilized: &markUtilized,
				CustomFields: customFields,
			}).Execute()

			if err != nil {
				return []model.Prefix{}, fmt.Errorf("failed to create prefix in Netbox: %v", err)
			}

			prefixes = append(prefixes, model.Prefix{
				PrefixID:    prefix.Id,
				Prefix:      ip + "/32",
				ExternalIPs: service.ExternalIPs,
				ServiceName: service.Name,
				Namespace:   service.Namespace,
			})
		}
	}

	return prefixes, nil
}

func (c *NetboxClient) DeletePrefix(id int32) error {
	_, err := c.netboxClient.IpamAPI.IpamPrefixesDestroy(context.Background(), id).Execute()
	return err
}

func NewNetboxClient(settings settings.Settings) (*NetboxClient, error) {
	client := netbox.NewAPIClientFor(settings.NetboxURL, settings.NetboxAPIToken)

	c := NetboxClient{
		netboxClient: client,
		settings:     settings,
	}
	return &c, nil
}
