{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "openfaas-sqs-connector.name" -}}
{{- .Chart.Name | default .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "openfaas-sqs-connector.fullname" -}}
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
Create a deployment name that's specific to the queue that it monitors.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "openfaas-sqs-connector.deployment-name" -}}

{{- $name := "" -}}
{{- if .name -}}
{{- $name = .name -}}
{{- else if .url -}}
{{- $name = regexSplit "/" .url -1 | last -}}
{{- end -}}

{{-/* TODO: make this prefix customizable based on the release name or name overrides */-}}
{{- $prefix := "openfaas-sqs-connector" -}}

{{- printf "%s-%s" $prefix $name | trunc 63 | trimSuffix "-" -}}

{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "openfaas-sqs-connector.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "openfaas-sqs-connector.labels" -}}
app.kubernetes.io/name: {{ include "openfaas-sqs-connector.name" . }}
helm.sh/chart: {{ include "openfaas-sqs-connector.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Queue CLI args
*/}}
{{- define "openfaas-sqs-connector.queue-args" -}}

{{ $result := list }}

{{- if .url -}}
  {{- $result = append $result "--queue-url" -}}
  {{- $result = append $result .url -}}

{{- else if .name }}
  {{- $result = append $result "--queue-name" -}}
  {{- $result = append $result .name -}}

  {{- if .awsAccountId -}}
    {{- $result = append $result "--aws-account-id" -}}
    {{- $result = append $result .awsAccountId -}}
  {{- end -}}

{{- else -}}
  {{ fail "Please provide either a queue name or URL" . }}
{{- end -}}

{{- if .region -}}
  {{- $result = append $result "--region" -}}
  {{- $result = append $result .region -}}
{{- end -}}

{{- $result = append $result "--max-number-of-messages" -}}
{{- $result = append $result (.maxNumberOfMessages | default 1) -}}
{{- $result = append $result "--max-wait-time" -}}
{{- $result = append $result (.maxWaitTime | default 1) -}}
{{- $result = append $result "--visibility-timeout" -}}
{{- $result = append $result (.visibilityTimeout | default 30) -}}

{{- range $result }}
- {{ . | quote }}
{{- end }}

{{- end -}}
