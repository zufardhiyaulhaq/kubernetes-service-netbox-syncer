package settings

import (
	"github.com/kelseyhightower/envconfig"
)

type Settings struct {
	NetboxAPIToken                    string              `envconfig:"NETBOX_API_TOKEN" required:"true"`
	NetboxURL                         string              `envconfig:"NETBOX_URL" required:"true"`
	KubernetesCluster                 string              `envconfig:"KUBERNETES_CLUSTER" default:"default"`
	KubernetesServiceAnnotationFilter []map[string]string `envconfig:"KUBERNETES_SERVICE_ANNOTATION_FILTER" default:""`
	KubernetesServiceLabelFilter      []map[string]string `envconfig:"KUBERNETES_SERVICE_LABEL_FILTER" default:""`
	KubernetesNamespaceFilter         []string            `envconfig:"KUBERNETES_NAMESPACE_FILTER" default:""`
	KubernetesTypeFilter              []string            `envconfig:"KUBERNETES_TYPE_FILTER" default:"LoadBalancer"`
}

func NewSettings() (Settings, error) {
	var settings Settings

	err := envconfig.Process("", &settings)
	if err != nil {
		return settings, err
	}

	return settings, nil
}
