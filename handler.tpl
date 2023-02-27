// Code generated by sqlc2oapi (https://github.com/bessiambre/sqlc2oapi). DO NOT EDIT.

package sqlcoa3api

import(
	"net/http"
	"github.com/ProlificLabs/snowball/sqlcapi/gen/sqlcoa3gen"
	"github.com/ProlificLabs/snowball/api/oa3api/middleware"
	"github.com/rotisserie/eris"
	"github.com/ProlificLabs/snowball/dcontext"
	"github.com/jackc/pgtype"
	"github.com/aarondl/opt/null"
	"github.com/ProlificLabs/snowball/sqlcapi/gen/apisqlc"
	"github.com/shopspring/decimal"
	"github.com/aarondl/chrono"
)

var _ apisqlc.Queries
var _ sqlcoa3gen.ErrHandled
var _ decimal.Decimal
var _ chrono.Date

{{ range $query := .Queries }}
/*** 
 * {{ .Name }}
 *
 * {{ .Text | replace "\n" "\n * "}}
 *
 * {{ .Cmd }}
 * {{ .Comments }}
 * {{ .Filename }}
 */
func (s *ServiceV3) {{ .Name }}(w http.ResponseWriter, r *http.Request{{ if gt (len .Params) 1 }}, body sqlcoa3gen.{{ .Name }}Inline{{ end }}) (*sqlcoa3gen.{{ .Name }}Return, error) {
	userId, _ := dcontext.UserID(r.Context())
	{{ if eq (len .Params) 1 }}
	res, err := s.Queries.{{ .Name }}(r.Context(){{ range .Params }}, {{ Oa3TypeTosqlcType .Column }}{{end}})
	{{- else }}
	res, err := s.Queries.{{ .Name }}(r.Context(), apisqlc.{{ .Name }}Params{
		{{ range .Params }}
			{{- snakeToGoCamel .Column.Name }}:{{ Oa3TypeTosqlcType .Column }},
		{{end -}}
		},
	)
	{{- end }}
	if err != nil {
		return nil, eris.Wrapf(err, "{{ .Name }}: query failure")
	}
	{{ if eq (len .Columns) 1 }}
	return &sqlcoa3gen.{{ .Name }}Return{
		{{- range .Columns }}
        {{ handlerReturnParamName . 0 }}: {{ sqlcTypeToOa3Type . 0 true}},
        {{- end }}
	}, nil
	{{- else }}
	return &sqlcoa3gen.{{ .Name }}Return{
		{{- range $i, $col := .Columns }}
        {{ handlerReturnParamName $col $i }}: {{ sqlcTypeToOa3Type $col $query.Name $i false}},
        {{- end }}
	}, nil
	{{- end }}
}
{{- end }}

func (s *ServiceV3) Wrap(next func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	return middleware.ErrorHandler(next)
}


func PgtypeJSONBtoMap(jsonb pgtype.JSONB) map[string]any {
	if jsonb.Get() == nil {
		return nil
	}
	switch v := jsonb.Get().(type) {
	case map[string]any:
		return v
	case pgtype.Status:
		return nil
	default:
		return nil
	}
}

func NullPgtypeJSONBtoMap(nulljsonb null.Val[pgtype.JSONB]) *map[string]any {
	if nulljsonb.IsNull(){
		return nil
	}

	jsonb:=nulljsonb.GetOrZero()

	if jsonb.Get() == nil {
		return nil
	}
	switch v := jsonb.Get().(type) {
	case map[string]any:
		return &v
	case pgtype.Status:
		return nil
	default:
		return nil
	}
}

func MapToPgtypeJSONB(in map[string]any) pgtype.JSONB {
	fd := pgtype.JSONB{Status: pgtype.Null}
	if in != nil {
		fd.Set(in)
	}
	return fd
}

func MapPtrToNullPgtypeJSONB(in *map[string]any) null.Val[pgtype.JSONB] {
	if in==nil{
		return null.From(pgtype.JSONB{Status: pgtype.Null})
	}
	fd := MapToPgtypeJSONB(*in)

	nullfd := null.FromCond(fd, false)
	if fd.Status == pgtype.Present {
		nullfd.Set(fd)
	}
	return nullfd
}


func PgtypeJSONtoMap(json pgtype.JSON) map[string]any {
	if json.Get() == nil {
		return nil
	}
	switch v := json.Get().(type) {
	case map[string]any:
		return v
	case pgtype.Status:
		return nil
	default:
		return nil
	}
}

func NullPgtypeJSONtoMap(nulljson null.Val[pgtype.JSON]) *map[string]any {
	if nulljson.IsNull(){
		return nil
	}

	json:=nulljson.GetOrZero()

	if json.Get() == nil {
		return nil
	}
	switch v := json.Get().(type) {
	case map[string]any:
		return &v
	case pgtype.Status:
		return nil
	default:
		return nil
	}
}

func MapToPgtypeJSON(in map[string]any) pgtype.JSON {
	fd := pgtype.JSON{Status: pgtype.Null}
	if in != nil {
		fd.Set(in)
	}
	return fd
}

func MapPtrToNullPgtypeJSON(in *map[string]any) null.Val[pgtype.JSON] {
	if in==nil{
		return null.From(pgtype.JSON{Status: pgtype.Null})
	}
	fd := MapToPgtypeJSON(*in)

	nullfd := null.FromCond(fd, false)
	if fd.Status == pgtype.Present {
		nullfd.Set(fd)
	}
	return nullfd
}