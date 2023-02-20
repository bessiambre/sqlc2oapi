package sqlcoa3api

import(
	"net/http"
	"github.com/ProlificLabs/snowball/sqlcapi/gen/sqlcoa3gen"
	"github.com/ProlificLabs/snowball/api/oa3api/middleware"
	"github.com/rotisserie/eris"
)

{{ range .Queries }}
/*** 
 * {{ .Name }}
 * {{ .Text | replace "\n" "\n * "}}
 * {{ .Cmd }}
 * {{ .Comments }}
 * {{ .Filename }}
 */
func (s *ServiceV3) {{ .Name }}(w http.ResponseWriter, r *http.Request{{ range .Params }}{{ sqlcToHandlerParam .Column }}{{end}}) (*sqlcoa3gen.{{ .Name }}Return, error) {
    res, err := s.Queries.{{ .Name }}(r.Context(){{ range .Params }}, {{ snakeToCamel .Column.Name }}{{end}})
	if err != nil {
		return nil, eris.Wrapf(err, "{{ .Name }}: query failure")
	}
	{{ if eq (len .Columns) 1 }}
	return &sqlcoa3gen.{{ .Name }}Return{
		{{- range .Columns }}
        {{ snakeToCamel .Name | title }}: res,
        {{- end }}
	}, nil
	{{- else }}
	return &sqlcoa3gen.{{ .Name }}Return{
		{{- range .Columns }}
        {{ snakeToCamel .Name | title }}: res.{{ snakeToCamel .Name | title }},
        {{- end }}
	}, nil
	{{- end }}
}
{{- end }}

func (s *ServiceV3) Wrap(next func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	return middleware.ErrorHandler(next)
}