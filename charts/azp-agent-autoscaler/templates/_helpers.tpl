{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "azp-agent-autoscaler.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "azp-agent-autoscaler.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create the name of the pod security policy to use
*/}}
{{- define "azp-agent-autoscaler.psp.name" -}}
{{- default .Values.rbac.psp.name (printf "%s-%s" .Release.Namespace (include "azp-agent-autoscaler.fullname" .)) | trunc 63 -}}
{{- end -}}

{{/*
Create the name of the pod security policy role(binding) to use
*/}}
{{- define "azp-agent-autoscaler.psp.rbacname" -}}
{{- printf "%s-psp" (include "azp-agent-autoscaler.fullname" . | trunc 59) -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "azp-agent-autoscaler.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "azp-agent-autoscaler.serviceAccountName" -}}
{{ .Values.serviceAccount.create | ternary (include "azp-agent-autoscaler.fullname" .) (.Values.serviceAccount.name | default "default") }}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "azp-agent-autoscaler.selector" -}}
name: {{ include "azp-agent-autoscaler.fullname" . }}
release: {{ .Release.Name }}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "azp-agent-autoscaler.labels" -}}
{{ include "azp-agent-autoscaler.selector" . }}
chart: {{ include "azp-agent-autoscaler.chart" . }}
{{- if .Chart.AppVersion }}
version: {{ .Chart.AppVersion | quote }}
{{- end }}
heritage: {{ .Release.Service }}
{{- end -}}

{{- define "azp-agent-autoscaler.stringDict" -}}
{{ range $key, $value := . }}
{{ $key | quote }}: {{ $value | quote }}
{{ end }}
{{- end -}}

{{/* https://github.com/openstack/openstack-helm-infra/blob/master/helm-toolkit/templates/utils/_joinListWithComma.tpl */}}
{{- define "helm-toolkit.utils.joinListWithComma" -}}
{{- $local := dict "first" true -}}
{{- range $k, $v := . -}}{{- if not $local.first -}},{{- end -}}{{- $v -}}{{- $_ := set $local "first" false -}}{{- end -}}
{{- end -}}
