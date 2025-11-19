.PHONY: readme
readme:
	helm-docs -c ./charts/kubernetes-service-netbox-syncer -d > README.md
	helm-docs -c ./charts/kubernetes-service-netbox-syncer

.PHONY: helm.create.releases
helm.create.releases:
	helm package charts/kubernetes-service-netbox-syncer --destination charts/releases
	helm repo index charts/releases