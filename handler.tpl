// Code generated by sqlc2oapi (https://github.com/bessiambre/sqlc2oapi). DO NOT EDIT.

package sqlcoa3api

import(
	"encoding/json"
	"net/http"

	"github.com/ProlificLabs/snowball/api/oa3api/middleware"
	"github.com/ProlificLabs/snowball/dcontext"
	"github.com/ProlificLabs/snowball/sqlcapi/gen/apisqlc"
	"github.com/ProlificLabs/snowball/sqlcapi/gen/sqlcoa3gen"
	"github.com/aarondl/chrono"
	"github.com/aarondl/opt/null"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
	"github.com/jackc/pgconn"
)

var _ apisqlc.Queries
var _ sqlcoa3gen.ErrHandled
var _ decimal.Decimal
var _ chrono.Date
var _ uuid.UUID

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
func (s *ServiceV3) {{ .Name }}(w http.ResponseWriter, r *http.Request{{ if gt (len .Params) 1 }}, body sqlcoa3gen.{{ .Name }}Params{{ end }}) ({{ if eq .Cmd ":one" }}*sqlcoa3gen.{{ .Name }}Return{{ else if eq .Cmd ":many" }}sqlcoa3gen.{{ .Name }}200Inline{{ end }}, error) {
	userId, _ := dcontext.UserID(r.Context())
	{{ if eq (len .Params) 1 }}
	pgRes, err := s.Queries.{{ .Name }}(r.Context(){{ range .Params }}, {{ Oa3TypeTosqlcType .Column $query.Name }}{{end}})
	{{- else }}
	pgRes, err := s.Queries.{{ .Name }}(r.Context(), apisqlc.{{ .Name }}Params{
		{{ range .Params }}
			{{- snakeToGoCamel .Column.Name }}:{{ Oa3TypeTosqlcType .Column $query.Name }},
		{{end -}}
		},
	)
	{{- end }}
	if err != nil {
		return nil, eris.Wrapf(processPgError(err), "{{ .Name }}; query failure")
	}

	{{ if eq .Cmd ":one" }}
	r := pgRes
	return &sqlcoa3gen.{{ .Name }}Return{
		{{- range $i, $col := .Columns }}
        {{ handlerReturnParamName $col $i }}: {{ sqlcTypeToOa3Type $col $query.Name $i ( eq (len $query.Columns) 1 ) }},
        {{- end }}
	}, nil
	{{ else if eq .Cmd ":many" }}
	resArray:=pgRes
	return func() sqlcoa3gen.{{ .Name }}200Inline {
		ret := make(sqlcoa3gen.{{ .Name }}200Inline,len(resArray))
		for _,r:= range resArray {
			ret=append(ret,sqlcoa3gen.{{ .Name }}Return{
			{{- range $i, $col := .Columns }}
        		{{ handlerReturnParamName $col $i }}: {{ sqlcTypeToOa3Type $col $query.Name $i ( eq (len $query.Columns) 1 ) }},
        	{{- end }}
			})
		}
		return ret;
	}(), nil

	{{ end }}
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

func MapToPgtypeJSONArray(in []map[string]any) []pgtype.JSON {
	out := make([]pgtype.JSON,len(in))
	for i:=range in{
		out[i]=MapToPgtypeJSON(in[i])
	}
	return out
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

func ParseUuid(s string) uuid.UUID {
	u, err := uuid.Parse(s)
	if err != nil {
		return uuid.UUID{}
	}
	return u
}

func UuidToStringArray(in []uuid.UUID) []string {
	out:=make([]string,len(in))
	for i:=range in{
		out[i]=in[i].URN()
	}
	return out
}

func ParseUuidArray(in []string) []uuid.UUID {
	out:=make([]uuid.UUID,len(in))
	for i:=range in{
		out[i]=ParseUuid(in[i])
	}
	return out
}

type PgJsonError struct {
	Message string `json:"message"`
	Code string `json:"code"`
}

func (e PgJsonError) Error() string {
	return e.Message
}

func processPgError(errIn error) error {
	errPg,ok:=errIn.(*pgconn.PgError)
	if !ok{
		return errIn
	}
	errOut := PgJsonError{}
	err := json.Unmarshal([]byte(errPg.Message), &errOut)
	if err != nil {
		return errIn
	}
	return errOut
}

func Int32ToInt16Array(in []int32)[]int16{
	out:=make([]int16,len(in))
	for i:=range in{
		out[i]=int16(in[i])
	}
	return out
}

func Int16ToInt32Array(in []int16)[]int32{
	out:=make([]int32,len(in))
	for i:=range in{
		out[i]=int32(in[i])
	}
	return out
}

func StringToBytesArray(in []string)[][]byte{
	out:=make([][]byte,len(in))
	for i:=range in{
		out[i]=[]byte(in[i])
	}
	return out
}

func BytesToStringArray(in [][]byte)[]string{
	out:=make([]string,len(in))
	for i:=range in{
		out[i]=string(in[i])
	}
	return out
}

func NullStringToBytes( in null.Val[string])[]byte{
	if in.IsSet(){
		return []byte(in.GetOrZero())
	}else{
		return nil
	}
}

func BytesToNullString( in []byte)null.Val[string]{
	if in != nil{
		return null.From(string(in))
	}else{
		return null.Val[string]{}
	}
}