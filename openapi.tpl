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
  {{ range .Queries }}
  ### {{ .Name }} ###
  #Query 
  #  {{ .Text | replace "\n" "\n  #  " }} 
  #Cmd {{ .Cmd }}
  #Comments {{ .Comments }}
  #Filename {{ .Filename }}
  /{{ camelSnake .Name }}:
    get:
      operationId: {{ .Name }}
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
              {{- range .Params }}
                {{ .Column.Name }}: {{ sqlToOa3Spec .Column }}
              {{- end }}
      responses:
        '200':
          description: Query result
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/{{ .Name }}Return'
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
    {{ .Name }}Return:
      type: object
      properties:
        {{- range $i , $col := .Columns }}
          {{ if $col.Name }}{{ $col.Name}}{{ else }}column{{ sum $i 1 }}{{ end }}: {{ sqlToOa3Spec . }}
        {{- end }}
      required:
        {{- range $i, $col  := .Columns }}
          - {{ if $col.Name }}{{ $col.Name}}{{ else }}column{{ sum $i 1 }}{{ end }}
        {{- end }}
  {{ end }}

security:
  - cookieAuth: []