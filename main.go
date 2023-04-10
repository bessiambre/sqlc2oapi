package main

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	ejson "encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"unicode"

	"github.com/Masterminds/sprig"
	"github.com/bessiambre/sqlc2oapi/pb"
)

//go:embed openapi.tpl
var openApiTpl string

//go:embed handler.tpl
var handlerTpl string

var TemplateFunctions = map[string]any{
	"camelSnake":             camelSnake,
	"snakeToCamel":           snakeToCamel,
	"snakeToGoCamel":         snakeToGoCamel,
	"sqlToOa3Spec":           sqlTypeToOa3SpecType,
	"sqlToHandlerParam":      sqlToHandlerParam,
	"sqlcTypeToOa3Type":      sqlcTypeToOa3Type,
	"handlerReturnParamName": handlerReturnParamName,
	"Oa3TypeTosqlcType":      Oa3TypeTosqlcType,
	"pathName":               pathName,
	"verbName":               verbName,
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error generating JSON: %s", err)
		os.Exit(2)
	}
}

func run() error {
	var req pb.CodeGenRequest
	reqBlob, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	if err := req.UnmarshalVT(reqBlob); err != nil {
		return err
	}

	GenerateJson(&req)

	resp, err := Generate(context.Background(), &req)
	if err != nil {
		return err
	}
	respBlob, err := resp.MarshalVT()
	if err != nil {
		return err
	}
	w := bufio.NewWriter(os.Stdout)
	if _, err := w.Write(respBlob); err != nil {
		return err
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return nil
}

func Generate(ctx context.Context, req *pb.CodeGenRequest) (*pb.CodeGenResponse, error) {
	// options, err := parseOptions(req)
	// if err != nil {
	// 	return nil, err
	// }

	filename := "codegen_oapi.yaml"
	// if options.Filename != "" {
	// 	filename = options.Filename
	// }

	//user_id is passed from logged in user don't include in oapi
	queriesByPathName := map[string][]*pb.Query{}
	var queriesForOapi []*pb.Query = make([]*pb.Query, 0, len(req.Queries))
	for _, query := range req.Queries {
		userIdFound := false
		newQuery := *query
		var newParams []*pb.Parameter = make([]*pb.Parameter, 0, len(query.Params))
		for _, param := range query.Params {
			if param.Column.Name != "user_id" {
				newParams = append(newParams, param)
			} else {
				userIdFound = true
			}
		}
		newQuery.Params = newParams
		queriesForOapi = append(queriesForOapi, &newQuery)

		if verbName(query.Name) == "" {
			return nil, fmt.Errorf("query name %s must start with http verb Get, Post, Put, Patch or Delete", query.Name)
		}

		queriesByPathName[pathName(query.Name)] = append(queriesByPathName[pathName(query.Name)], &newQuery)

		if !userIdFound {
			return nil, fmt.Errorf("query %s must take @user_id parameter to verify permissions", query.Name)
		}
	}

	tmpl, err := template.New("openapi").Funcs(sprig.FuncMap()).Funcs(TemplateFunctions).Parse(openApiTpl)
	if err != nil {
		return nil, err
	}
	buff := new(bytes.Buffer)
	err = tmpl.Execute(buff, map[string]any{
		"QueriesByPathName": queriesByPathName,
		"Queries":           queriesForOapi,
	})
	if err != nil {
		return nil, err
	}

	filenameHandlers := "sqlcoa3api/handlers.go"
	tmplHandlers, err := template.New("handlers").Funcs(sprig.FuncMap()).Funcs(TemplateFunctions).Parse(handlerTpl)
	if err != nil {
		return nil, err
	}
	buffHandlers := new(bytes.Buffer)
	err = tmplHandlers.Execute(buffHandlers, map[string]any{
		"Queries": req.Queries,
	})
	if err != nil {
		return nil, err
	}

	return &pb.CodeGenResponse{
		Files: []*pb.File{
			{
				Name:     filename,
				Contents: append(buff.Bytes(), '\n'),
			},
			{
				Name:     filenameHandlers,
				Contents: append(buffHandlers.Bytes(), '\n'),
			},
		},
	}, nil
}

func parseOptions(req *pb.CodeGenRequest) (*pb.JSONCode, error) {
	if req.Settings == nil {
		return new(pb.JSONCode), nil
	}
	if req.Settings.Codegen != nil {
		if len(req.Settings.Codegen.Options) != 0 {
			var options *pb.JSONCode
			dec := ejson.NewDecoder(bytes.NewReader(req.Settings.Codegen.Options))
			dec.DisallowUnknownFields()
			if err := dec.Decode(&options); err != nil {
				return options, fmt.Errorf("unmarshalling options: %s", err)
			}
			return options, nil
		}
	}
	if req.Settings.Json != nil {
		return req.Settings.Json, nil
	}
	return new(pb.JSONCode), nil
}

// schema_UserIDProfile -> schema_user_id_profile
// ID -> id
func camelSnake(filename string) string {
	build := new(strings.Builder)

	var upper bool

	in := []rune(filename)
	for i, r := range []rune(in) {
		if !unicode.IsLetter(r) {
			upper = false
			build.WriteRune(r)
			continue
		}

		if !unicode.IsUpper(r) {
			upper = false
			build.WriteRune(r)
			continue
		}

		addUnderscore := false
		if upper {
			if i+1 < len(in) && unicode.IsLower(in[i+1]) {
				addUnderscore = true
			}
		} else {
			if i-1 > 0 && unicode.IsLetter(in[i-1]) {
				addUnderscore = true
			}
		}

		if addUnderscore {
			build.WriteByte('_')
		}

		upper = true
		build.WriteRune(unicode.ToLower(r))
	}

	return build.String()
}

func snakeToCamel(in string) string {
	build := new(strings.Builder)
	sawUnderscore := false
	for _, r := range in {
		if r == '_' {
			sawUnderscore = true
			continue
		}

		if sawUnderscore {
			sawUnderscore = false
			build.WriteRune(unicode.ToUpper(r))
		} else {
			build.WriteRune(r)
		}
	}

	return build.String()
}

// func snakeToCamel(name string) string {
// 	out := new(strings.Builder)
// 	for _, p := range strings.Split(name, "_") {
// 		out.WriteString(strings.Title(p))
// 	}

// 	return out.String()
// }

func snakeToGoCamel(name string) string {
	out := new(strings.Builder)
	for _, p := range strings.Split(name, "_") {
		if p == "id" { // matches sqlc's func StructName
			out.WriteString("ID")
		} else {
			out.WriteString(strings.Title(p))
		}
	}

	return out.String()
}

func schemaAndName(in pb.Identifier) (string, string) {
	if in.Schema != "" {
		return in.Schema, in.Name
	} else {
		s := strings.Split(in.Name, ".")
		if len(s) == 2 {
			return s[0], s[1]
		} else {
			return in.Schema, in.Name
		}
	}
}

func sqlTypeToOa3SpecType(in *pb.Column) string {
	typeStr := "type: object"

	schema, name := schemaAndName(*in.Type)
	if schema != "pg_catalog" && schema != "" && name != "citext" {
		//assume it's an enum for now
		typeStr = "type: string"
	} else {

		switch in.Type.Name {
		case "int2", "pg_catalog.int2":
			typeStr = "type: integer, format: int32"
		case "int4", "pg_catalog.int4":
			typeStr = "type: integer, format: int32"
		case "int8", "pg_catalog.int8":
			typeStr = "type: integer, format: int64"
		case "numeric", "pg_catalog.numeric", "money", "pg_catalog.money":
			typeStr = "type: string, format: decimal"
		case "float4", "pg_catalog.float4":
			typeStr = "type: number, format: float"
		case "float8", "pg_catalog.float8":
			typeStr = "type: number, format: double"
		case "text", "varchar", "bpchar", "pg_catalog.text", "pg_catalog.varchar", "pg_catalog.bpchar":
			typeStr = "type: string"
		case "bytea", "pg_catalog.bytea":
			typeStr = "type: string, format: byte"
		case "timestamp", "pg_catalog.timestamp":
			typeStr = "type: string, format: date-time"
		case "timestamptz", "pg_catalog.timestamptz":
			typeStr = "type: string, format: date-time"
		case "date", "pg_catalog.date":
			typeStr = "type: string, format: date"
		case "time", "pg_catalog.time":
			typeStr = "type: string, format: time"
		case "bool", "pg_catalog.bool":
			typeStr = "type: boolean"
		case "jsonb", "pg_catalog.jsonb":
			typeStr = "type: object"
		case "json", "pg_catalog.json":
			typeStr = "type: object"
		case "citext":
			typeStr = "type: string"
		case "uuid":
			typeStr = "type: string"
		}
	}
	if !in.NotNull {
		typeStr += ", nullable: true"
	}

	if in.IsArray {
		return fmt.Sprintf("{ type: array, items: { %s } }", typeStr)
	} else {
		return fmt.Sprintf("{ %s } # %s %s", typeStr, in.Type.Schema, in.Type.Name)
	}
}

func sqlToHandlerParam(in *pb.Column) string {

	//skip userid as it is extracted from ctx
	if in.Name == "user_id" {
		return ""
	}

	typeStr := "any"

	typeSchema, typeName := schemaAndName(*in.Type)
	if typeSchema != "pg_catalog" && typeSchema != "" && typeName != "citext" {
		//assume it's an enum for now
		typeStr = "string"
	} else {

		switch in.Type.Name {
		case "int2", "pg_catalog.int2":
			typeStr = "int32"
		case "int4", "pg_catalog.int4":
			typeStr = "int32"
		case "int8", "pg_catalog.int8":
			typeStr = "int64"
		case "numeric", "pg_catalog.numeric", "money", "pg_catalog.money":
			typeStr = "decimal.Decimal"
		case "float4", "pg_catalog.float4":
			typeStr = "float64"
		case "float8", "pg_catalog.float8":
			typeStr = "float64"
		case "text", "varchar", "bpchar", "pg_catalog.text", "pg_catalog.varchar", "pg_catalog.bpchar":
			typeStr = "string"
		case "bytea", "pg_catalog.bytea":
			typeStr = "[]byte"
		case "timestamptz", "pg_catalog.timestamptz":
			typeStr = "chrono.DateTime"
		case "timestamp", "pg_catalog.timestamp":
			typeStr = "chrono.DateTime"
		case "date", "pg_catalog.date":
			typeStr = "chrono.Date"
		case "time", "pg_catalog.time":
			typeStr = "chrono.Time"
		case "bool", "pg_catalog.bool":
			typeStr = "bool"
		case "jsonb", "pg_catalog.jsonb":
			typeStr = "map[string]any"
		case "json", "pg_catalog.json":
			typeStr = "map[string]any"
		case "citext":
			typeStr = "string"
		case "uuid":
			typeStr = "string"
		}
	}

	if !in.NotNull {
		typeStr = fmt.Sprintf("null.Val[%s]", typeStr)
	}

	if in.IsArray {
		return fmt.Sprintf(", %s []%s", snakeToCamel(in.Name), typeStr)
	} else {
		return fmt.Sprintf(", %s %s", snakeToCamel(in.Name), typeStr)
	}
}

func sqlcTypeToOa3Type(in *pb.Column, queryName string, i int, single bool) string {
	convStr := ""

	name := in.Name
	if in.Name == "" {
		name = "Column" + strconv.Itoa((i + 1))
	}

	varName := "res"
	if !single {
		varName = "res." + strings.Title(snakeToGoCamel(name))
	}

	typeSchema, typeName := schemaAndName(*in.Type)
	if typeSchema != "pg_catalog" && typeSchema != "" && typeName != "citext" {
		//assume it's an enum for now
		if in.NotNull {
			return "string(" + varName + ")"
		} else {
			return "null.FromCond(string(" + varName + "." + strings.Title(snakeToGoCamel(name)) + "), " + varName + ".Valid)"
		}
	}

	switch in.Type.Name {
	case "int2", "pg_catalog.int2":
		if in.IsArray {
			convStr = "Int16ToInt32Array(" + varName + ")"
		} else if in.NotNull {
			convStr = "int32(" + varName + ")"
		} else {
			convStr = "null.FromCond(int32(" + varName + ".GetOrZero())," + varName + ".IsSet())"
		}
	case "json", "pg_catalog.json":
		if in.IsArray {
			typeName := queryName + "Return" + strings.Title(snakeToCamel(name)) + "Item"
			convStr = `
			func(in []pgtype.JSON)[]sqlcoa3gen.` + typeName + `{
				out:=make([]sqlcoa3gen.` + typeName + `,len(in))
				for i:=range in{
					out[i]=PgtypeJSONtoMap(in[i])
				}
				return out
			}(` + varName + `)`
		} else if in.NotNull {
			convStr = "(sqlcoa3gen." + queryName + "Return" + strings.Title(snakeToCamel(name)) + ")(PgtypeJSONtoMap(" + varName + "))"
		} else {
			convStr = "(*sqlcoa3gen." + queryName + "Return" + strings.Title(snakeToCamel(name)) + ")(NullPgtypeJSONtoMap(" + varName + "))"
		}
	case "jsonb", "pg_catalog.jsonb":
		if in.IsArray {
			typeName := queryName + "Return" + strings.Title(snakeToCamel(name)) + "Item"
			convStr = `
			func(in []pgtype.JSONB)[]sqlcoa3gen.` + typeName + `{
				out:=make([]sqlcoa3gen.` + typeName + `,len(in))
				for i:=range in{
					out[i]=PgtypeJSONBtoMap(in[i])
				}
				return out
			}(` + varName + `)`
		} else if in.NotNull {
			convStr = "(sqlcoa3gen." + queryName + "Return" + strings.Title(snakeToCamel(name)) + ")(PgtypeJSONBtoMap(" + varName + "))"
		} else {
			convStr = "(*sqlcoa3gen." + queryName + "Return" + strings.Title(snakeToCamel(name)) + ")(NullPgtypeJSONBtoMap(" + varName + "))"
		}
	case "numeric", "pg_catalog.numeric", "money", "pg_catalog.money":
		if in.NotNull {
			convStr = varName
		} else {
			convStr = "null.FromCond(" + varName + ".Decimal, " + varName + ".Valid)"
		}
	case "uuid", "pg_catalog.uuid":
		if in.NotNull {
			convStr = varName + ".URN()"
		} else {
			convStr = "null.FromCond(" + varName + ".UUID.URN(), " + varName + ".Valid)"
		}
	case "bytea", "pg_catalog.bytea":
		if in.IsArray {
			convStr = "BytesToStringArray(" + varName + ")"
		} else if in.NotNull {
			convStr = "string(" + varName + ")"
		} else {
			convStr = "BytesToNullString(" + varName + ")"
		}
	default:
		convStr = varName
	}

	return convStr
}

func Oa3TypeTosqlcType(in *pb.Column) string {
	varName := "body." + strings.Title(snakeToCamel(in.Name))
	if in.Name == "user_id" {
		return "userId"
	}

	typeSchema, typeName := schemaAndName(*in.Type)
	if typeSchema != "pg_catalog" && typeSchema != "" && typeName != "citext" {
		//assume it's an enum for now
		if in.NotNull {
			return "(apisqlc." + strings.Title(snakeToGoCamel(typeName)) + ")(" + varName + ")"
		} else {
			return "apisqlc.Null" + strings.Title(snakeToGoCamel(typeName)) + "(" + strings.Title(snakeToGoCamel(in.Name)) + ":apisqlc." + strings.Title(snakeToGoCamel(in.Name)) + "(" + varName + ".GetOrZero()), Valid:" + varName + ".IsSet())"
		}
	}

	convStr := ""
	switch in.Type.Name {
	case "int2", "pg_catalog.int2":
		if in.IsArray {
			convStr = "Int32ToInt16Array(" + varName + ")"
		} else if in.NotNull {
			convStr = "int16(" + varName + ")"
		} else {
			convStr = "null.FromCond(int16(" + varName + ".GetOrZero())," + varName + ".IsSet())"
		}
	case "json", "pg_catalog.json":
		if in.IsArray {
			convStr = "MapToPgtypeJSONArray(([]map[string]any)(" + varName + "))"
		} else if in.NotNull {
			convStr = "MapToPgtypeJSON(" + varName + ")"
		} else {
			convStr = "MapPtrToNullPgtypeJSON((*map[string]any)(" + varName + "))"
		}
	case "jsonb", "pg_catalog.jsonb":
		if in.IsArray {
			convStr = "MapToPgtypeJSONBArray(([]map[string]any)(" + varName + "))"
		} else if in.NotNull {
			convStr = "MapToPgtypeJSONB(" + varName + ")"

		} else {
			convStr = "MapPtrToNullPgtypeJSONB((*map[string]any)(" + varName + "))"
		}
	case "numeric", "pg_catalog.numeric", "money", "pg_catalog.money":
		if in.NotNull {
			convStr = varName
		} else {
			convStr = "decimal.NullDecimal{Decimal:" + varName + ".GetOrZero(), Valid:" + varName + ".IsSet()}"
		}
	case "uuid", "pg_catalog.uuid":
		if in.IsArray {
			convStr = "ParseUuidArray(" + varName + ")"
		} else if in.NotNull {
			convStr = "ParseUuid(" + varName + ")"
		} else {
			convStr = "uuid.NullUUID{UUID:ParseUuid(" + varName + ".GetOrZero()), Valid:" + varName + ".IsSet()}"
		}
	case "bytea", "pg_catalog.bytea":
		if in.IsArray {
			convStr = "StringToBytesArray(" + varName + ")"
		} else if in.NotNull {
			convStr = "[]byte(" + varName + ")"
		} else {
			convStr = "NullStringToBytes(" + varName + ")"
		}
	default:
		convStr = varName
	}
	return convStr
}

func handlerReturnParamName(in *pb.Column, index int) string {
	if in.Name != "" {
		return strings.Title(snakeToCamel(in.Name))
	}
	return "Column" + strconv.Itoa(index+1)
}

func pathName(operationName string) string {
	if strings.HasPrefix(operationName, "Get") {
		return camelSnake(operationName[3:])
	} else if strings.HasPrefix(operationName, "Put") {
		return camelSnake(operationName[3:])
	} else if strings.HasPrefix(operationName, "Patch") {
		return camelSnake(operationName[5:])
	} else if strings.HasPrefix(operationName, "Post") {
		return camelSnake(operationName[4:])
	} else if strings.HasPrefix(operationName, "Delete") {
		return camelSnake(operationName[6:])
	} else {
		return camelSnake(operationName)
	}
}

func verbName(operationName string) string {
	if strings.HasPrefix(operationName, "Get") {
		return "get"
	} else if strings.HasPrefix(operationName, "Put") {
		return "put"
	} else if strings.HasPrefix(operationName, "Patch") {
		return "patch"
	} else if strings.HasPrefix(operationName, "Post") {
		return "post"
	} else if strings.HasPrefix(operationName, "Delete") {
		return "delete"
	} else {
		return ""
	}
}

// for debugging
func GenerateJson(req *pb.CodeGenRequest) error {
	bytes, err := json.MarshalIndent(req, "", "\t")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(os.TempDir(), "sqlcPluginInput.json"), bytes, 0o644)
	// os.Stdout.Write(b)
	return nil
}
