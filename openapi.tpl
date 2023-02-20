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
  #Text {{ .Text | replace "\n" "\n #" }} 
  #Cmd {{ .Cmd }}
  #Comments {{ .Comments }}
  #Filename {{ .Filename }}
  /{{ .Name }}:
    parameters:
    {{- range .Params }}
      - name: {{ .Column.Name }}
        in: query
        required: true
        schema: {{ sqlcToOa3Spec .Column }} #${{ .Number }}
    {{- end }}
    get:
      operationId: {{ .Name }}
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
        {{- range .Columns }}
          {{ .Name }}: {{ sqlcToOa3Spec . }}
        {{- end }}
      required:
        {{- range .Columns }}
          - {{ .Name }}
        {{- end }}
  {{ end }}

security:
  - cookieAuth: []