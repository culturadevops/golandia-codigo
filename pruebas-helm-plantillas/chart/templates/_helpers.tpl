{{- define "web.nombre" -}}
{{ .Release.Name }}-{{ .Chart.Name }}
{{- end }}

{{- define "web.etiquetas" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
