package client

import (
	"github.com/netbox-community/go-netbox/v4"
	"github.com/zufardhiyaulhaq/kubernetes-service-netbox-syncer/settings"
)

type NetboxClient struct {
	netboxClient *netbox.APIClient
}

func (c *NetboxClient) Client() *netbox.APIClient {
	return c.netboxClient
}

func NewNetboxClient(settings settings.Settings) (*NetboxClient, error) {
	client := netbox.NewAPIClientFor(settings.NetboxURL, settings.NetboxAPIToken)

	c := NetboxClient{
		netboxClient: client,
	}
	return &c, nil
}
