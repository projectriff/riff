{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "riff.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "riff.fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create the list of kafka broker nodes to use
*/}}
{{- define "riff.kafkaBrokers" -}}
{{- if .Values.kafka.create -}}
    {{ default ( printf "%s-%s.%s:9092" .Release.Name "kafka" .Release.Namespace ) .Values.kafka.broker.nodes }}
{{- else -}}
    {{ .Values.kafka.broker.nodes }}
{{- end -}}
{{- end -}}

{{/*
Create the list of kafka zookeeper nodes to use
*/}}
{{- define "riff.kafkaZkNodes" -}}
{{- if .Values.kafka.create -}}
    {{ default ( printf "%s-%s.%s:2181" .Release.Name "zookeeper" .Release.Namespace ) .Values.kafka.zookeeper.nodes }}
{{- else -}}
    {{ .Values.kafka.zookeeper.nodes }}
{{- end -}}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "riff.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "riff.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}
