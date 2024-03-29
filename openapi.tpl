openapi: 3.0.0
info:
  title: Pulley sqlc API
  version: 0.0.1
  description: OpenAPI spec for the sqlc Pulley API
servers:
  - url: '{environment}/v3'
    variables:
      environment:
        default: https://api.pulley.com
        enum:
          - https://api.pulley.com
          - https://api-master.pulley.com
          - http://localhost:8080
paths:
{{ range $key, $queries := .QueriesByPathName -}}
  ### {{ $key }} ###
  /{{ pathName $key }}:
  {{ range $queries -}}
  #Query 
  #  {{ .Text | replace "\n" "\n  #  " }} 
  #Cmd {{ .Cmd }}
  #Comments {{ .Comments }}
  #Filename {{ .Filename }}
    {{ verbName .Name }}:
      operationId: {{ .Name }}
      {{- if .Params }}
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/{{ .Name }}Params'
              
      {{- end }}
      responses:
        '200':
          description: Query result
          content:
            application/json:
              schema:
              {{ if eq .Cmd ":one" }}
                $ref: '#/components/schemas/{{ .Name }}Return'
              {{ else if eq .Cmd ":many" }}
                { type: array, items: { $ref: '#/components/schemas/{{ .Name }}Return' } }
              {{ end }}
  {{ end }}
{{ end }}

components:
  securitySchemes:
    cookieAuth: # arbitrary name for the security scheme; will be used in the "security" key later
      type: apiKey
      in: cookie
      name: pulley-development # cookie name

### BEGIN parameters ###
  #parameters:

### BEGIN schemas ###
  schemas:
  {{- range .Queries }}
  {{- if .Params }}
    {{ .Name }}Params:
      type: object
      properties:
      {{- range .Params }}
        {{ .Column.Name }}: {{ sqlToOa3Spec .Column }}
      {{- end }}
      required:
      {{- range .Params }}
        - {{ .Column.Name }}
      {{- end -}}
  {{- end }}

    {{ .Name }}Return:
      type: object
      properties:
      {{- range $i , $col := .Columns }}
        {{ if $col.Name }}{{ $col.Name}}{{ else }}column{{ add $i 1 }}{{ end }}: {{ sqlToOa3Spec . }}
      {{- end }}
      required:
      {{- range $i, $col  := .Columns }}
        - {{ if $col.Name }}{{ $col.Name}}{{ else }}column{{ add $i 1 }}{{ end }}
      {{- end }}
  {{ end }}

security:
  - cookieAuth: []