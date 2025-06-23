{{- define "agentry.name" -}}
agentry
{{- end -}}

{{- define "agentry.fullname" -}}
{{ include "agentry.name" . }}
{{- end -}}
