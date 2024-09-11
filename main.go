package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/ettle/strcase"
	"github.com/sqlc-dev/sqlc/internal/codegen/sdk"
	"github.com/sqlc-dev/sqlc/internal/plugin"

	"embed"

	"google.golang.org/protobuf/proto"
)

//go:embed templates/*
var templates embed.FS

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
	MessageName     string
	EnumName        string
}

const (
	OptGenerate = ":generate"
	OptPackage  = ":package"
	OptReplace  = ":replace"
)

const (
	templateDir    = "templates/"
	headerTmplName = "header.tmpl"
	headerTmplCode = templateDir + headerTmplName
	msgTmplName    = "message.tmpl"
	msgTmplCode    = templateDir + msgTmplName
	enumTmplName   = "enum.tmpl"
	enumTmplCode   = templateDir + enumTmplName

	wkprefix = "google.protobuf."

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
	WellKnownDecimal          = wkprefix + "Decimal"
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
	WellKnownMoney            = wkprefix + "Money"
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

	// imports
	ImportGoogleProtobuf              = "google/protobuf/"
	ImportGoogleProtobufAPI           = ImportGoogleProtobuf + "api"
	ImportGoogleProtobufAny           = ImportGoogleProtobuf + "any"
	ImportGoogleProtobufDecimal       = ImportGoogleProtobuf + "decimal"
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
		WellKnownDecimal:          ImportGoogleProtobufDecimal,
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

	// headerTmpl = template.Must(template.New(headerTmplCode).
	// ParseFiles(headerTmplCode))
	headerTmpl = template.Must(template.New("header.tmpl").
			ParseFS(templates, headerTmplCode))
	msgTmpl = template.Must(template.New("message.tmpl").
		ParseFS(templates, msgTmplCode))
	enumTmpl = template.Must(template.New("enum.tmpl").
			ParseFS(templates, enumTmplCode))
)

type protoType struct {
	ImportPath string
	Name       string
}

func getGenerateOpts(comments []string) (*generateOpts, error) {
	replace := make(map[string]protoType)
	g := &generateOpts{
		Replace: replace,
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
			g.Generate = true
		}

		for _, cmdOption := range []string{
			"package",
			"replace",
			"outdir",
			"filename",
			"messagename",
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
			case "messagename":
				if len(part) != 3 {
					return nil, fmt.Errorf(
						"-- messagename: <messagename>... takes exactly 1 argument",
					)
				}
				msgName := part[2]
				g.MessageName = msgName
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

type headerInput struct {
	PackageName string
	Imports     map[string]string
}

type msgInput struct {
	Name   string
	Fields fields
}

type enumInput struct {
	Name   string
	Values []string
}

type protofiles map[string]*protofile

type message struct {
	name   string
	fields map[string]*field
}

type enum struct {
	name   string
	values []string
}

type protofile struct {
	headerBuf       *bytes.Buffer
	msgsBuf         *bytes.Buffer
	enumsBuf        *bytes.Buffer
	messages        map[string]*message
	enums           map[string]*enum
	folderPath      string
	packagename     string
	filename        string
	requiredImports map[string]string
	fullpath        string
}

func pkgToPath(s string) string {
	return strings.ReplaceAll(s, ".", "/")
}

// func addToMsgProtoFiles(comments []string, columns *)
func (g *generator) appendProtos(
	input interface{},
) error {
	switch v := input.(type) {
	case *plugin.Table:
		o, err := getGenerateOpts(v.RawComments)
		if !o.Generate {
			break
		}
		if err != nil {
			return err
		}
		if o.MessageName == "" {
			o.MessageName = protoName(v.Rel.Name)
		}
		opts, err := setOptDefaults(o, false, "message.proto")
		if err != nil {
			return err
		}
		err = g.assembleProto(opts, v.Columns, nil)
		if err != nil {
			return err
		}
	case *plugin.Query:
		o, err := getGenerateOpts(v.RawComments)
		if !o.Generate {
			break
		}
		if err != nil {
			return err
		}
		opts, err := setOptDefaults(o, true, "messsage.proto")
		if err != nil {
			return err
		}
		err = g.assembleProto(opts, v.Columns, nil)
		if err != nil {
			return err
		}
	case *plugin.Enum:
		o, err := getGenerateOpts(v.RawComments)
		if !o.Generate {
			break
		}
		if err != nil {
			return err
		}
		if o.EnumName == "" {
			o.EnumName = protoName(v.Name)
		}
		opts, err := setOptDefaults(o, false, "enum.proto")
		if err != nil {
			return err
		}
		err = g.assembleProto(opts, nil, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *generator) assembleProto(
	opts *generateOpts,
	columns []*plugin.Column,
	e *plugin.Enum,
) error {
	if !opts.Generate {
		return nil
	}
	pf := g.buildProtoFile(opts)
	if e != nil {
		g.handleEnum(opts, e)
	}
	if len(columns) > 0 {
		g.handleMsg(opts, columns, pf)
	}

	return nil
}

func setOptDefaults(
	opts *generateOpts,
	query bool,
	filename string,
) (*generateOpts, error) {
	// Handle Empty Options with Defaults
	if opts.Package == "" {
		opts.Package = "sqlcgen"
	}
	if opts.OutputDirectory == "" {
		opts.OutputDirectory = "sqlcgen"
	}
	if opts.FileName == "" {
		opts.FileName = filename
	}
	if query && opts.MessageName == "" {
		return nil, fmt.Errorf(
			"Queries that want to generate protobufs must declare a messagename: <messagename>",
		)
	}

	return opts, nil
}

func (g *generator) buildProtoFile(opts *generateOpts) *protofile {
	// Build File
	header := bytes.Buffer{}
	msgs := bytes.Buffer{}
	enums := bytes.Buffer{}
	folderpath := opts.OutputDirectory + "/" + pkgToPath(opts.Package)
	pf := &protofile{
		headerBuf:       &header,
		msgsBuf:         &msgs,
		enumsBuf:        &enums,
		folderPath:      folderpath,
		packagename:     opts.Package,
		filename:        opts.FileName,
		fullpath:        folderpath + "/" + opts.FileName,
		requiredImports: make(map[string]string),
		messages:        make(map[string]*message),
		enums:           make(map[string]*enum),
	}

	if g.protofiles[opts.Package] == nil {
		g.protofiles[opts.Package] = pf
	}

	return pf
}

func (g *generator) handleImports(c *plugin.Column, opts *generateOpts, pf *protofile) string {
	replace, ok := opts.Replace[c.Name]
	if ok {
		if requiredImport(replace.Name) {
			pf.requiredImports[replace.Name] = replace.ImportPath
		}
		return replace.Name
	}

	// If we don't find a type it may be because its an enum
	ptype := pgTypeToProtoType(c)
	if ptype == "unknown" {
		pn := protoName(c.Type.Name)
		for name, pro := range g.protofiles {
			if g.protofiles[name].enums[pn] != nil {
				ptype = name + "." + pn
				pf.requiredImports[ptype] = pkgToPath(name) + "/" + pro.filename
			}
		}
		return ptype
	}

	if requiredImport(ptype) {
		pf.requiredImports[ptype] = protoTypeMap[ptype]
	}

	return ptype
}

func (g *generator) handleMsg(opts *generateOpts, columns []*plugin.Column, pf *protofile) {
	opts.MessageName = protoName(opts.MessageName)

	if g.protofiles[opts.Package].messages[opts.MessageName] == nil {
		// Build Messsages
		msg := &message{
			name:   opts.MessageName,
			fields: make(map[string]*field),
		}
		g.protofiles[opts.Package].messages[opts.MessageName] = msg
	}

	for _, c := range columns {
		f := field{}
		f.PrimaryKey = c.PrimaryKey
		f.IsArray = c.IsArray
		f.Name = fieldName(c.Name)
		// protoType := pf.pgTypeToProtoType(c)
		f.Type = g.handleImports(c, opts, pf)

		// Build Field
		if g.protofiles[opts.Package].messages[opts.MessageName].fields[f.Name] == nil {
			g.protofiles[opts.Package].messages[opts.MessageName].fields[f.Name] = &f
		}
	}
}

func (g *generator) handleEnum(opts *generateOpts, e *plugin.Enum) {
	opts.EnumName = protoName(opts.EnumName)
	if g.protofiles[opts.Package].enums[opts.EnumName] == nil {
		// Build Messsages
		en := &enum{
			name:   opts.EnumName,
			values: enumName(opts.EnumName, e.Vals),
		}
		g.protofiles[opts.Package].enums[opts.EnumName] = en
	}

}

func enumName(n string, s []string) []string {
	u := strcase.ToSNAKE(n)
	x := []string{u + "_UNSPECIFIED"}
	for _, z := range s {
		y := strcase.ToSNAKE(z)
		x = append(x, (u + "_" + y))
	}
	return x
}

type generator struct {
	protofiles map[string]*protofile
}

func run() error {
	var req plugin.GenerateRequest
	reqBlob, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(reqBlob, &req); err != nil {
		return err
	}

	protofiles := make(map[string]*protofile)
	g := &generator{
		protofiles: protofiles,
	}

	// Schema must be first because we only append columns to the map[string]*field
	// queries that don't already exist.  Fields on *plugin.Column aren't
	// parsed in queries.
	schemas := req.Catalog.GetSchemas()
	for _, s := range schemas {
		if s.Name == "pg_catalog" || s.Name == "information_schema" {
			continue
		}

		// Enums must go first,  because schema tables will match against
		// those types
		enums := s.GetEnums()
		for _, e := range enums {
			if err := g.appendProtos(e); err != nil {
				return err
			}
		}

		tables := s.GetTables()
		for _, t := range tables {
			if err := g.appendProtos(t); err != nil {
				return err
			}
		}

	}

	// Queries can be added to the .proto,  Column data like "primarykey"
	queries := req.GetQueries()
	for _, query := range queries {
		if err := g.appendProtos(query); err != nil {
			return err
		}

	}

	for pkg, output := range protofiles {
		header := output.headerBuf
		msgs := output.msgsBuf
		enums := output.enumsBuf

		if err := headerTmpl.Execute(header, &headerInput{
			PackageName: pkg,
			Imports:     output.requiredImports,
		}); err != nil {
			return err
		}

		for name, message := range output.messages {
			// Need maps to be determinsitic
			sortedFields := make(fields, 0, len(message.fields))
			for _, f := range message.fields {
				sortedFields = append(sortedFields, *f)
			}
			sort.Sort(sortedFields)

			if err := msgTmpl.Execute(msgs, &msgInput{
				Name:   name,
				Fields: sortedFields,
			}); err != nil {
				return err
			}
		}

		for name, e := range output.enums {
			if err := enumTmpl.Execute(enums, &enumInput{
				Name:   name,
				Values: e.values,
			}); err != nil {
				return err
			}
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
		if len(output.messages) > 0 {
			_, err = f.Write(msgs.Bytes())
			if err != nil {
				return err
			}
		}

		// Append Messages to .protofile
		if len(output.enums) > 0 {
			_, err = f.Write(enums.Bytes())
			if err != nil {
				return err
			}
		}

		if err := f.Close(); err != nil {
			return err
		}
	}

	return nil
}

type field struct {
	Name       string
	Type       string
	IsArray    bool
	PrimaryKey bool
}

type fields []field

func (f fields) Len() int {
	return len(f)
}

func (f fields) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

func (f fields) Less(i, j int) bool {
	// Always put "primarykey" first
	if f[i].PrimaryKey {
		return true
	}
	if f[j].PrimaryKey {
		return false
	}
	// Otherwise, sort alphabetically by Name
	return f[i].Name < f[j].Name
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

// Type, NotFound
func pgTypeToProtoType(col *plugin.Column) string {
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

	case "numeric", "pg_catalog.numeric":
		return WellKnownDecimal

	case "money":
		return WellKnownMoney

	case "boolean", "bool", "pg_catalog.bool":
		if notNull {
			return protoBool
		}
		return WellKnownBoolValue

	case "json":
		return WellKnownStruct

	case "uuid", "jsonb", "bytea", "blob", "pg_catalog.bytea":
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
		return "unknown"
	}

	return "unknown"
}
