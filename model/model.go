package model

type SyncResult struct {
	PrefixID    int64
	Prefix      string
	DNS         string
	ServiceName string
	Namespace   string
}

type KubernetesService struct {
	Name        string
	Namespace   string
	ExternalIPs string
}
