package main

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	ejson "encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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
	"camelSnake":              camelSnake,
	"snakeToCamel":            snakeToCamel,
	"snakeToGoCamel":          snakeToGoCamel,
	"sqlToOa3Spec":            sqlTypeToOa3SpecType,
	"sqlToHandlerParam":       sqlToHandlerParam,
	"sqlcTypeToOa3Type":       sqlcTypeToOa3Type,
	"sqlcTypeToOa3TypeSingle": sqlcTypeToOa3TypeSingle,
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
	userIdFound := false
	var queriesForOapi []*pb.Query = make([]*pb.Query, 0, len(req.Queries))
	for _, query := range req.Queries {
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
	}

	if !userIdFound {
		return nil, errors.New("query must take @user_id parameter to verify permissions")
	}

	tmpl, err := template.New("openapi").Funcs(sprig.FuncMap()).Funcs(TemplateFunctions).Parse(openApiTpl)
	if err != nil {
		return nil, err
	}
	buff := new(bytes.Buffer)
	err = tmpl.Execute(buff, map[string]any{
		"Queries": queriesForOapi,
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

// func snakeToCamel(in string) string {
// 	build := new(strings.Builder)
// 	sawUnderscore := false
// 	for _, r := range in {
// 		if r == '_' {
// 			sawUnderscore = true
// 			continue
// 		}

// 		if sawUnderscore {
// 			sawUnderscore = false
// 			build.WriteRune(unicode.ToUpper(r))
// 		} else {
// 			build.WriteRune(r)
// 		}
// 	}

// 	return build.String()
// }

func snakeToCamel(name string) string {
	out := new(strings.Builder)
	for _, p := range strings.Split(name, "_") {
		out.WriteString(strings.Title(p))
	}

	return out.String()
}

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

func sqlTypeToOa3SpecType(in *pb.Column) string {
	typeStr := "type: object"

	if in.Type.Schema != "pg_catalog" {
		//assume it's an enum for now
		typeStr = "type: string"
	} else {

		switch in.Type.Name {
		case "int4", "pg_catalog.int4":
			typeStr = "type: integer, format: int32"
		case "numeric", "pg_catalog.numeric":
			typeStr = "type: string, format: decimal"
		case "text":
			typeStr = "type: string"
		case "date":
			typeStr = "type: string, format: date"
		case "timestamptz":
			typeStr = "type: string, format: date-time"
		case "bool":
			typeStr = "type: boolean"
		case "jsonb":
			typeStr = "type: object"
		case "json":
			typeStr = "type: object"
		}
	}
	if !in.NotNull {
		typeStr += ", nullable: true"
	}

	return fmt.Sprintf("{ %s }", typeStr)
}

func sqlToHandlerParam(in *pb.Column) string {

	//skip userid as it is extracted from ctx
	if in.Name == "user_id" {
		return ""
	}

	typeStr := "any"

	if in.Type.Schema != "pg_catalog" {
		//assume it's an enum for now
		typeStr = "string"
	} else {

		switch in.Type.Name {
		case "int4", "pg_catalog.int4":
			typeStr = "int32"
		case "numeric", "pg_catalog.numeric":
			typeStr = "decimal.Decimal"
		case "text":
			typeStr = "string"
		case "date":
			typeStr = "chrono.Date"
		case "timestamptz":
			typeStr = "chrono.DateTime"
		case "bool", "pg_catalog.bool":
			typeStr = "bool"
		case "jsonb":
			typeStr = "map[string]any"
		case "json":
			typeStr = "map[string]any"
		}
	}

	if !in.NotNull {
		typeStr = fmt.Sprintf("null.Val[%s]", typeStr)
	}

	return fmt.Sprintf(", %s %s", snakeToCamel(in.Name), typeStr)
}

func sqlcTypeToOa3Type(in *pb.Column, queryName string) string {
	convStr := ""

	if in.Type.Schema != "pg_catalog" {
		//assume it's an enum for now
		return in.Type.Schema + "string(res." + strings.Title(snakeToGoCamel(in.Name)) + ")"
	}

	switch in.Type.Name {
	case "jsonb":
		if in.NotNull {
			convStr = "(*sqlcoa3gen." + queryName + "Return" + strings.Title(snakeToCamel(in.Name)) + ")(PgtypeJSONBtoMap(res." + strings.Title(snakeToGoCamel(in.Name)) + "))"
		} else {
			convStr = "(*sqlcoa3gen." + queryName + "Return" + strings.Title(snakeToCamel(in.Name)) + ")(NullPgtypeJSONBtoMap(res." + strings.Title(snakeToGoCamel(in.Name)) + "))"
		}
	default:
		convStr = "res." + strings.Title(snakeToGoCamel(in.Name))
	}

	return convStr
}

func sqlcTypeToOa3TypeSingle(in *pb.Column, queryName string) string {
	convStr := ""

	switch in.Type.Name {
	case "jsonb":
		if !in.NotNull {
			convStr = "PgtypeJSONBtoMap(res)"
		} else {
			convStr = "PgtypeJSONBtoMap(res)"
		}
	default:
		convStr = "res"
	}

	return convStr
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
