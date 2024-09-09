package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/ettle/strcase"
	"github.com/sqlc-dev/sqlc/internal/codegen/sdk"
	"github.com/sqlc-dev/sqlc/internal/plugin"

	// "github.com/ryboe/q"

	// "google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error generating JSON: %s", err)
		os.Exit(2)
	}
}

type generateOpts struct {
	Generate        bool
	Package         string
	Replace         map[string]protoType
	FileName        string
	OutputDirectory string
}

const (
	OptGenerate = ":generate"
	OptPackage  = ":package"
	OptReplace  = ":replace"
)

const (
	msgTmplCode    = "message.tmpl"
	headerTmplCode = "header.tmpl"
	wkprefix       = "google.protobuf."

	// proto represents builtin types
	protoDouble   = "double"
	protoFloat    = "float"
	protoInt32    = "int32"
	protoInt64    = "int64"
	protoUint32   = "uint32"
	protoUint64   = "uint64"
	protoSint32   = "sint32"
	protoSint64   = "sint64"
	protoFixed32  = "fixed32"
	protoFixed64  = "fixed64"
	protoSFixed32 = "sfixed32"
	protoSFixed64 = "sfixed64"
	protoBool     = "bool"
	protoString   = "string"
	protoBytes    = "bytes"

	// well known represents google.proto. well known types
	WellKnownAny              = wkprefix + "Any"
	WellKnownBoolValue        = wkprefix + "BoolValue"
	WellKnownBytesValue       = wkprefix + "BytesValue"
	WellKnownDoubleValue      = wkprefix + "DoubleValue"
	WellKnownDuration         = wkprefix + "Duration"
	WellKnownEmpty            = wkprefix + "Empty"
	WellKnownEnum             = wkprefix + "Enum"
	WellKnownEnumValue        = wkprefix + "EnumValue"
	WellKnownField            = wkprefix + "Field"
	WellKnownFieldCardinality = wkprefix + "Field.Cardinality"
	WellKnownFieldKind        = wkprefix + "Field.Kind"
	WellKnownFieldMask        = wkprefix + "FieldMask"
	WellKnownFloatValue       = wkprefix + "FloatValue"
	WellKnownInt32Value       = wkprefix + "Int32Value"
	WellKnownInt64Value       = wkprefix + "Int64Value"
	WellKnownListValue        = wkprefix + "ListValue"
	WellKnownMethod           = wkprefix + "Method"
	WellKnownMixin            = wkprefix + "Mixin"
	WellKnownNullValue        = wkprefix + "NullValue"
	WellKnownOption           = wkprefix + "Option"
	WellKnownSourceContext    = wkprefix + "SourceContext"
	WellKnownStringValue      = wkprefix + "StringValue"
	WellKnownStruct           = wkprefix + "Struct"
	WellKnownSyntax           = wkprefix + "Syntax"
	WellKnownTimestamp        = wkprefix + "Timestamp"
	WellKnownType             = wkprefix + "Type"
	WellKnownUInt32Value      = wkprefix + "UInt32Value"
	WellKnownUInt64Value      = wkprefix + "UInt64Value"
	WellKnownValue            = wkprefix + "Value"
	WellKnownMoney            = wkprefix + "Money"

	// imports
	ImportGoogleProtobuf              = "google/protobuf/"
	ImportGoogleProtobufAny           = ImportGoogleProtobuf + "any"
	ImportGoogleProtobufAPI           = ImportGoogleProtobuf + "api"
	ImportGoogleProtobufDuration      = ImportGoogleProtobuf + "duration"
	ImportGoogleProtobufEmpty         = ImportGoogleProtobuf + "empty"
	ImportGoogleProtobufEnum          = ImportGoogleProtobuf + "type"
	ImportGoogleProtobufEnumValue     = ImportGoogleProtobuf + "type"
	ImportGoogleProtobufFieldMask     = ImportGoogleProtobuf + "field_mask"
	ImportGoogleProtobufListValue     = ImportGoogleProtobuf + "struct"
	ImportGoogleProtobufMoney         = ImportGoogleProtobuf + "money"
	ImportGoogleProtobufSourceContext = ImportGoogleProtobuf + "source_context"
	ImportGoogleProtobufStruct        = ImportGoogleProtobuf + "struct"
	ImportGoogleProtobufTimestamp     = ImportGoogleProtobuf + "timestamp"
	ImportGoogleProtobufType          = ImportGoogleProtobuf + "type"
	ImportGoogleProtobufValue         = ImportGoogleProtobuf + "struct"
	ImportGoogleProtobufWrappers      = ImportGoogleProtobuf + "wrappers"
)

var (
	protoTypeMap = map[string]string{
		WellKnownAny:              ImportGoogleProtobufAny,
		WellKnownBoolValue:        ImportGoogleProtobufWrappers,
		WellKnownBytesValue:       ImportGoogleProtobufWrappers,
		WellKnownDoubleValue:      ImportGoogleProtobufWrappers,
		WellKnownDuration:         ImportGoogleProtobufDuration,
		WellKnownEmpty:            ImportGoogleProtobufEmpty,
		WellKnownEnum:             ImportGoogleProtobufEnum,
		WellKnownEnumValue:        ImportGoogleProtobufEnumValue,
		WellKnownField:            ImportGoogleProtobufType,
		WellKnownFieldCardinality: ImportGoogleProtobufType,
		WellKnownFieldKind:        ImportGoogleProtobufType,
		WellKnownFieldMask:        ImportGoogleProtobufFieldMask,
		WellKnownFloatValue:       ImportGoogleProtobufWrappers,
		WellKnownInt32Value:       ImportGoogleProtobufWrappers,
		WellKnownInt64Value:       ImportGoogleProtobufWrappers,
		WellKnownListValue:        ImportGoogleProtobufStruct,
		WellKnownMethod:           ImportGoogleProtobufAPI,
		WellKnownMixin:            ImportGoogleProtobufAPI,
		WellKnownNullValue:        ImportGoogleProtobufStruct,
		WellKnownOption:           ImportGoogleProtobufType,
		WellKnownSourceContext:    ImportGoogleProtobufSourceContext,
		WellKnownStringValue:      ImportGoogleProtobufWrappers,
		WellKnownStruct:           ImportGoogleProtobufStruct,
		WellKnownSyntax:           ImportGoogleProtobufType,
		WellKnownTimestamp:        ImportGoogleProtobufTimestamp,
		WellKnownType:             ImportGoogleProtobufType,
		WellKnownUInt32Value:      ImportGoogleProtobufWrappers,
		WellKnownUInt64Value:      ImportGoogleProtobufWrappers,
		WellKnownValue:            ImportGoogleProtobufStruct,
		WellKnownMoney:            ImportGoogleProtobufMoney,
	}

	headerTmpl = template.Must(template.New(headerTmplCode).
			ParseFiles(headerTmplCode))
	msgTmpl = template.Must(template.New(msgTmplCode).
		ParseFiles(msgTmplCode))
)

type protoType struct {
	ImportPath string
	Name       string
}

func getGenerateOpts(comments []string) (*generateOpts, error) {
	// params := make(map[string]string)
	// flags := make(map[string]bool)
	// var cleanedComments []string
	replace := make(map[string]protoType)
	g := &generateOpts{
		Generate: true,
		Replace:  replace,
	}

	for _, line := range comments {
		var prefix string
		if strings.HasPrefix(line, "--") {
			prefix = "--"
		}
		if strings.HasPrefix(line, "/*") {
			prefix = "/*"
		}
		if strings.HasPrefix(line, "#") {
			prefix = "#"
		}
		if prefix == "" {
			continue
		}
		rest := line[len(prefix):]
		if !strings.Contains(rest, ":") {
			continue
		}
		for _, flagOpt := range []string{
			"generate",
		} {
			if !strings.HasPrefix(strings.TrimSpace(rest), flagOpt) {
				continue
			}
			opt := fmt.Sprintf(" %s:", flagOpt)

			if !strings.HasPrefix(rest, opt) {
				return nil, fmt.Errorf("invalid metadata: %s", line)
			}
		}
		g.Generate = true

		for _, cmdOption := range []string{
			"package",
			"replace",
			"outdir",
			"filename",
		} {
			if !strings.HasPrefix(strings.TrimSpace(rest), cmdOption) {
				continue
			}
			opt := fmt.Sprintf(" %s: ", cmdOption)

			if !strings.HasPrefix(rest, opt) {
				return nil, fmt.Errorf("invalid metadata: %s", line)
			}

			part := strings.Split(strings.TrimSpace(line), " ")

			switch cmdOption {
			case "package":
				if len(part) != 3 {
					return nil, fmt.Errorf("-- package: <package>... takes exactly 1 argument")
				}
				packageName := part[2]
				g.Package = packageName
			case "outdir":
				if len(part) != 3 {
					return nil, fmt.Errorf("-- outdir: <outdir>... takes exactly 1 argument")
				}
				outdir := part[2]
				g.OutputDirectory = outdir
			case "filename":
				if len(part) != 3 {
					return nil, fmt.Errorf("-- filename: <filename>... takes exactly 1 argument")
				}
				fileName := part[2]
				g.FileName = fileName
			case "replace":
				if len(part) != 5 {
					return nil, fmt.Errorf(
						"-- replace: <column> <import> <type> ... takes exactly 3 argument",
					)
				}
				columnName := part[2]
				importPath := part[3]
				typeName := part[4]
				g.Replace[columnName] = protoType{
					ImportPath: importPath,
					Name:       typeName,
				}
			}
		}
	}

	return g, nil
}

type HeaderInput struct {
	PackageName string
	Imports     map[string]string
}

type protofile struct {
	header          *bytes.Buffer
	msgs            *bytes.Buffer
	folderPath      string
	packagename     string
	filename        string
	requiredImports map[string]string
}

func pkgToPath(s string) string {
	return strings.ReplaceAll(s, ".", "/")
}

func run() error {
	protofiles := make(map[string]protofile)
	var req plugin.GenerateRequest

	reqBlob, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}

	if err := proto.Unmarshal(reqBlob, &req); err != nil {
		return err
	}

	schemas := req.Catalog.GetSchemas()
	for _, s := range schemas {
		if s.Name == "pg_catalog" || s.Name == "information_schema" {
			continue
		}

		tables := s.GetTables()
		for _, t := range tables {
			opts, err := getGenerateOpts(t.RawComments)
			if err != nil {
				return err
			}
			if !opts.Generate {
				continue
			}
			if opts.Package == "" {
				opts.Package = "sqlcgen"
			}

			if opts.FileName == "" {
				opts.FileName = "message.proto"
			}
			if opts.OutputDirectory == "" {
				opts.OutputDirectory = "sqlcgen"
			}

			msgs := bytes.Buffer{}
			header := bytes.Buffer{}
			file := protofile{
				folderPath:      opts.OutputDirectory + "/" + pkgToPath(opts.Package),
				packagename:     opts.Package,
				filename:        opts.FileName,
				msgs:            &msgs,
				header:          &header,
				requiredImports: make(map[string]string),
			}

			msgInput := file.toMsgInput(t, opts)
			if err := msgTmpl.Execute(&msgs, msgInput); err != nil {
				return err
			}

			protofiles[opts.Package] = file
		}
	}

	for pkg, output := range protofiles {
		header := output.header
		msgs := output.msgs
		if err := headerTmpl.Execute(header, &HeaderInput{
			PackageName: pkg,
			Imports:     output.requiredImports,
		}); err != nil {
			return err
		}

		if _, err := os.Stat(output.folderPath); err != nil {
			if err := os.MkdirAll(output.folderPath, os.ModePerm); err != nil {
				return err
			}
		}

		// If the file doesn't exist, create it, or append to the file
		f, err := os.Create(output.folderPath + "/" + output.filename)
		if err != nil {
			return err
		}

		//  Append Syntax, Package, []import
		_, err = f.Write(header.Bytes())
		if err != nil {
			return err
		}

		// Append Messages to .protofile
		_, err = f.Write(msgs.Bytes())
		if err != nil {
			return err
		}

		if err := f.Close(); err != nil {
			return err
		}
	}

	return nil
}

type msgInput struct {
	Name   string
	Fields []Field
}

type Field struct {
	Name    string
	Type    string
	IsArray bool
}

func protoName(s string) string {
	r := strings.ToLower(s)
	r = strcase.ToPascal(s)

	return r
}

func fieldName(s string) string {
	s = strings.ToLower(s)
	s = strcase.ToSnake(s)
	return s
}

func (g *protofile) toMsgInput(tbl *plugin.Table, genOpts *generateOpts) msgInput {
	m := msgInput{Name: protoName(tbl.Rel.Name)}
	for _, c := range tbl.Columns {
		f := Field{}

		f.IsArray = c.IsArray
		f.Name = fieldName(c.Name)
		protoType := g.pgTypeToProtoType(c)

		replace, ok := genOpts.Replace[f.Name]
		if ok {
			if requiredImport(replace.Name) {
				g.requiredImports[replace.Name] = replace.ImportPath
			}
			protoType = replace.Name
		} else {
			if requiredImport(protoType) {
				g.requiredImports[protoType] = protoTypeMap[protoType]
			}
		}
		f.Type = protoType

		m.Fields = append(m.Fields, f)
	}

	return m
}

func requiredImport(s string) bool {
	switch s {
	case protoDouble:
		return false
	case protoFloat:
		return false
	case protoInt32:
		return false
	case protoInt64:
		return false
	case protoUint32:
		return false
	case protoUint64:
		return false
	case protoSint32:
		return false
	case protoSint64:
		return false
	case protoFixed32:
		return false
	case protoFixed64:
		return false
	case protoSFixed32:
		return false
	case protoSFixed64:
		return false
	case protoBool:
		return false
	case protoString:
		return false
	case protoBytes:
		return false
	}

	return true
}

func (g *protofile) pgTypeToProtoType(col *plugin.Column) string {
	columnType := sdk.DataType(col.Type)
	notNull := col.NotNull || col.IsArray
	// driver := parseDriver(options./*  */SqlPackage)

	columnType = strings.ToLower(columnType)

	switch columnType {
	// Int32
	case "integer",
		"int",
		"int4",
		"pg_catalog.int4",
		"serial",
		"serial4",
		"pg_catalog.serial4",
		"smallserial",
		"smallint", "int2", "pg_catalog.int2", "serial2",
		"pg_catalog.serial2":
		if notNull {
			return protoInt32
		}
		return WellKnownInt32Value

	// Int64
	case "interval",
		"pg_catalog.interval",
		"bigint",
		"int8",
		"pg_catalog.int8",
		"bigserial",
		"serial8",
		"pg_catalog.serial8":
		if notNull {
			return protoInt64
		}
		return WellKnownInt64Value

	// Float
	case "real",
		"float4",
		"pg_catalog.float4",
		"float",
		"double precision",
		"float8",
		"pg_catalog.float8":
		if notNull {
			return protoFloat
		}
		return WellKnownFloatValue

	case "numeric", "pg_catalog.numeric", "money":
		return WellKnownMoney

	case "boolean", "bool", "pg_catalog.bool":
		if notNull {
			return protoBool
		}
		return WellKnownBoolValue

	case "json":
		return WellKnownStruct

	case "jsonb":
		if notNull {
			return protoBytes
		}
		return WellKnownBytesValue

	case "bytea", "blob", "pg_catalog.bytea":
		if notNull {
			return protoBytes
		}
		return WellKnownBytesValue

	case "pg_catalog.timestamptz",
		"date",
		"timestamptz",
		"pg_catalog.timestamp",
		"pg_catalog.timetz",
		"pg_catalog.time":
		return WellKnownTimestamp

	case "citext",
		"lquery",
		"ltree",
		"ltxtquery",
		"name",
		"inet",
		"cidr",
		"macaddr",
		"macaddr8",
		"pg_catalog.bpchar",
		"pg_catalog.varchar",
		"string",
		"text":
		if notNull {
			return protoString
		}
		return WellKnownStringValue

	case "uuid":
		if notNull {
			return protoBytes
		}
		return WellKnownBytesValue

	// All these PG Range Types Required FieldOptions
	// Handle this Later
	case "daterange":
		// switch driver {
		// case opts.SQLDriverPGXV4:
		// 	return "pgtype.Daterange"
		// case opts.SQLDriverPGXV5:
		// 	return "pgtype.Range[pgtype.Date]"
		// default:
		// 	return "interface{}"
		// }

	case "datemultirange":
		// switch driver {
		// case opts.SQLDriverPGXV5:
		// 	return "pgtype.Multirange[pgtype.Range[pgtype.Date]]"
		// default:
		// 	return "interface{}"
		// }

	case "tsrange":
		// switch driver {
		// case opts.SQLDriverPGXV4:
		// 	return "pgtype.Tsrange"
		// case opts.SQLDriverPGXV5:
		// 	return "pgtype.Range[pgtype.Timestamp]"
		// default:
		// 	return "interface{}"
		// }

	case "tsmultirange":
		// switch driver {
		// case opts.SQLDriverPGXV5:
		// 	return "pgtype.Multirange[pgtype.Range[pgtype.Timestamp]]"
		// default:
		// 	return "interface{}"
		// }

	case "tstzrange":
		// switch driver {
		// case opts.SQLDriverPGXV4:
		// 	return "pgtype.Tstzrange"
		// case opts.SQLDriverPGXV5:
		// 	return "pgtype.Range[pgtype.Timestamptz]"
		// default:
		// 	return "interface{}"
		// }

	case "tstzmultirange":
		// switch driver {
		// case opts.SQLDriverPGXV5:
		// 	return "pgtype.Multirange[pgtype.Range[pgtype.Timestamptz]]"
		// default:
		// 	return "interface{}"
		// }

	case "numrange":
		// switch driver {
		// case opts.SQLDriverPGXV4:
		// 	return "pgtype.Numrange"
		// case opts.SQLDriverPGXV5:
		// 	return "pgtype.Range[pgtype.Numeric]"
		// default:
		// 	return "interface{}"
		// }

	case "nummultirange":
		// switch driver {
		// case opts.SQLDriverPGXV5:
		// 	return "pgtype.Multirange[pgtype.Range[pgtype.Numeric]]"
		// default:
		// 	return "interface{}"
		// }

	case "int4range":
		// switch driver {
		// case opts.SQLDriverPGXV4:
		// 	return "pgtype.Int4range"
		// case opts.SQLDriverPGXV5:
		// 	return "pgtype.Range[pgtype.Int4]"
		// default:
		// 	return "interface{}"
		// }

	case "int4multirange":
		// switch driver {
		// case opts.SQLDriverPGXV5:
		// 	return "pgtype.Multirange[pgtype.Range[pgtype.Int4]]"
		// default:
		// 	return "interface{}"
		// }

	case "int8range":
		// switch driver {
		// case opts.SQLDriverPGXV4:
		// 	return "pgtype.Int8range"
		// case opts.SQLDriverPGXV5:
		// 	return "pgtype.Range[pgtype.Int8]"
		// default:
		// 	return "interface{}"
		// }

	case "int8multirange":
		// switch driver {
		// case opts.SQLDriverPGXV5:
		// 	return "pgtype.Multirange[pgtype.Range[pgtype.Int8]]"
		// default:
		// 	return "interface{}"
		// }

	case "hstore":
		// if driver.IsPGX() {
		// 	return "pgtype.Hstore"
		// }
		return WellKnownAny

	case "bit", "varbit", "pg_catalog.bit", "pg_catalog.varbit":
		// if driver == opts.SQLDriverPGXV5 {
		// 	return "pgtype.Bits"
		// }
		// if driver == opts.SQLDriverPGXV4 {
		// 	return "pgtype.Varbit"
		// }

	case "cid":
		// if driver == opts.SQLDriverPGXV5 {
		// 	return "pgtype.Uint32"
		// }
		// if driver == opts.SQLDriverPGXV4 {
		// 	return "pgtype.CID"
		// }

	case "oid":
		// if driver == opts.SQLDriverPGXV5 {
		// 	return "pgtype.Uint32"
		// }
		// if driver == opts.SQLDriverPGXV4 {
		// 	return "pgtype.OID"
		// }

	case "tid":
		// if driver.IsPGX() {
		// 	return "pgtype.TID"
		// }

	case "xid":
		// if driver == opts.SQLDriverPGXV5 {
		// 	return "pgtype.Uint32"
		// }
		// if driver == opts.SQLDriverPGXV4 {
		// 	return "pgtype.XID"
		// }

	case "box":
		// if driver.IsPGX() {
		// 	return "pgtype.Box"
		// }

	case "circle":
		// if driver.IsPGX() {
		// 	return "pgtype.Circle"
		// }

	case "line":
		// if driver.IsPGX() {
		// 	return "pgtype.Line"
		// }

	case "lseg":
		// if driver.IsPGX() {
		// 	return "pgtype.Lseg"
		// }

	case "path":
		// if driver.IsPGX() {
		// 	return "pgtype.Path"
		// }

	case "point":
		// if driver.IsPGX() {
		// 	return "pgtype.Point"
		// }

	case "polygon":
		// if driver.IsPGX() {
		// 	return "pgtype.Polygon"
		// }

	case "vector":
		// if driver == opts.SQLDriverPGXV5 {
		// 	if emitPointersForNull {
		// 		return "*pgvector.Vector"
		// 	} else {
		// 		return "pgvector.Vector"
		// 	}
		// }
		return "WARNING:not-implemented"

	case "void":
		// A void value can only be scanned into an empty interface.
		return WellKnownAny

	case "any":
		return WellKnownAny

	default:
		return WellKnownAny
	}

	return WellKnownAny
}
