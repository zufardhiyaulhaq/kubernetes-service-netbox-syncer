{{- define "kubernetes-service-netbox-syncer.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "kubernetes-service-netbox-syncer.podlabels" -}}
{{- with .Values.podLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}
