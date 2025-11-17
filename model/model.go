package model

type Prefix struct {
	PrefixID    int32  `json:"prefix_id"`
	Prefix      string `json:"prefix"`
	ExternalIPs string `json:"dns"`
	ServiceName string `json:"service_name"`
	Namespace   string `json:"namespace"`
}

type KubernetesService struct {
	Name        string
	Namespace   string
	ExternalIPs string
}
