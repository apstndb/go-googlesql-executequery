package catalog

import (
	"fmt"

	googlesql "github.com/goccy/go-googlesql"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// Demo proto schema used by the sample catalog.
//
// Implementation note: the descriptor-loading
// path used here (`descriptorpb.FileDescriptorProto` → `proto.Marshal`
// → `googlesql.NewFileDescriptorProto.ParseFromString` →
// `DescriptorPool.BuildFile`) IS the natural binding of the upstream
// C++ idiom: `google::protobuf::DescriptorPool::BuildFile(
// const FileDescriptorProto&)` takes wire-format bytes deserialised
// into a FileDescriptorProto, and that is exactly what this code
// does. The choice we made is upstream of that — we *hand-build* the
// FileDescriptorProto via `descriptorpb` (no protoc dev-tool
// dependency), keeping the schema in Go source so the demo is
// self-contained. If upstream parity ever becomes a goal, swap this
// out for embedded descriptor bytes generated from
// `googlesql/testdata/*.proto` via protoc.
//
// The package name and a representative subset of fields mirror
// upstream `sample_catalog_impl.cc` so analyzer-test queries that
// reference `zetasql_test.TestEnum` or `zetasql_test.KitchenSinkPB`
// resolve through the same names.
const (
	sampleProtoPackage  = "zetasql_test"
	sampleProtoFileName = "zetasql_test/sample.proto"
	sampleProtoMsg      = "KitchenSinkPB"
	sampleProtoEnum     = "TestEnum"
)

// sampleProtoFullName returns the fully-qualified proto name (the
// "package.LocalName" form go-googlesql expects when looking up types).
func sampleProtoFullName(local string) string {
	return sampleProtoPackage + "." + local
}

// buildSampleFileDescriptorWire produces the wire-format
// FileDescriptorProto for the demo schema. Kept private so callers
// only see the resulting `*googlesql.DescriptorPool`.
func buildSampleFileDescriptorWire() ([]byte, error) {
	fileName := sampleProtoFileName
	syntax := "proto2"
	pkg := sampleProtoPackage
	enumName := sampleProtoEnum
	msgName := sampleProtoMsg
	enumZero, enumOne, enumTwo := "TESTENUM0", "TESTENUM1", "TESTENUM2"
	zero, one, two := int32(0), int32(1), int32(2)
	int64Field, stringField, enumField := "int64_val", "string_val", "test_enum"
	int64Tag, stringTag, enumTag := int32(1), int32(2), int32(3)
	int64Type := descriptorpb.FieldDescriptorProto_TYPE_INT64
	stringType := descriptorpb.FieldDescriptorProto_TYPE_STRING
	enumType := descriptorpb.FieldDescriptorProto_TYPE_ENUM
	optional := descriptorpb.FieldDescriptorProto_LABEL_OPTIONAL
	enumTypeRef := "." + sampleProtoFullName(sampleProtoEnum)

	fdp := &descriptorpb.FileDescriptorProto{
		Name:    &fileName,
		Package: &pkg,
		Syntax:  &syntax,
		EnumType: []*descriptorpb.EnumDescriptorProto{{
			Name: &enumName,
			Value: []*descriptorpb.EnumValueDescriptorProto{
				{Name: &enumZero, Number: &zero},
				{Name: &enumOne, Number: &one},
				{Name: &enumTwo, Number: &two},
			},
		}},
		MessageType: []*descriptorpb.DescriptorProto{{
			Name: &msgName,
			Field: []*descriptorpb.FieldDescriptorProto{
				{Name: &int64Field, Number: &int64Tag, Type: &int64Type, Label: &optional},
				{Name: &stringField, Number: &stringTag, Type: &stringType, Label: &optional},
				// Marshalling an Enum field needs Type=TYPE_ENUM and
				// TypeName pointing at the enum's fully-qualified name
				// (with a leading "."), per the descriptor.proto contract.
				{Name: &enumField, Number: &enumTag, Type: &enumType, Label: &optional, TypeName: &enumTypeRef},
			},
		}},
	}
	wire, err := proto.Marshal(fdp)
	if err != nil {
		return nil, fmt.Errorf("marshal sample FileDescriptorProto: %w", err)
	}
	return wire, nil
}

// sampleProtoPostBuild attaches a DescriptorPool with the demo
// schema, registers the proto/enum types under user-friendly names,
// and adds TestTable + EnumTable as proto/enum-typed sample tables.
//
// Mirrored into schema.Tables for DESCRIBE, but those entries are
// purely descriptive — the catalog's SimpleTable for each is built
// directly here and go-googlesql clears the handle on AddOwnedTable.
func sampleProtoPostBuild(cat *googlesql.SimpleCatalog, schema *Schema, _ map[string]*googlesql.SimpleTable, tf *googlesql.TypeFactory) error {
	wire, err := buildSampleFileDescriptorWire()
	if err != nil {
		return err
	}

	pool, err := googlesql.NewDescriptorPool()
	if err != nil {
		return fmt.Errorf("new descriptor pool: %w", err)
	}
	wfdp, err := googlesql.NewFileDescriptorProto()
	if err != nil {
		return fmt.Errorf("new wasm FileDescriptorProto: %w", err)
	}
	if ok, err := wfdp.ParseFromString(string(wire)); err != nil {
		return fmt.Errorf("ParseFromString: %w", err)
	} else if !ok {
		return fmt.Errorf("ParseFromString returned false on a well-formed payload")
	}
	if _, err := pool.BuildFile(wfdp); err != nil {
		return fmt.Errorf("BuildFile: %w", err)
	}

	msgFQN := sampleProtoFullName(sampleProtoMsg)
	enumFQN := sampleProtoFullName(sampleProtoEnum)
	msgDesc, err := pool.FindMessageTypeByName(msgFQN)
	if err != nil {
		return fmt.Errorf("FindMessageTypeByName %q: %w", msgFQN, err)
	}
	if msgDesc == nil {
		return fmt.Errorf("FindMessageTypeByName %q returned nil", msgFQN)
	}
	enumDesc, err := pool.FindEnumTypeByName(enumFQN)
	if err != nil {
		return fmt.Errorf("FindEnumTypeByName %q: %w", enumFQN, err)
	}
	if enumDesc == nil {
		return fmt.Errorf("FindEnumTypeByName %q returned nil", enumFQN)
	}

	if err := cat.SetDescriptorPool(pool); err != nil {
		return fmt.Errorf("SetDescriptorPool: %w", err)
	}

	protoType, err := tf.MakeProtoType(msgDesc, []string{sampleProtoPackage, sampleProtoMsg})
	if err != nil {
		return fmt.Errorf("MakeProtoType: %w", err)
	}
	enumType, err := tf.MakeEnumType(enumDesc, []string{sampleProtoPackage, sampleProtoEnum})
	if err != nil {
		return fmt.Errorf("MakeEnumType: %w", err)
	}

	int32Type, err := tf.MakeSimpleType(googlesql.TypeKindTypeInt32)
	if err != nil {
		return fmt.Errorf("make INT32 type: %w", err)
	}

	// TestTable: key INT32, TestEnum, KitchenSink (proto).
	if err := addSampleProtoTable(cat, "TestTable", []sampleColumn{
		{name: "key", typ: int32Type},
		{name: "TestEnum", typ: enumType},
		{name: "KitchenSink", typ: protoType},
	}); err != nil {
		return err
	}
	// EnumTable: key INT32, TestEnum.
	if err := addSampleProtoTable(cat, "EnumTable", []sampleColumn{
		{name: "key", typ: int32Type},
		{name: "TestEnum", typ: enumType},
	}); err != nil {
		return err
	}

	// Mirror into the Go-side Schema so DESCRIBE prints the columns.
	protoLabel := "PROTO<" + msgFQN + ">"
	enumLabel := "ENUM<" + enumFQN + ">"
	schema.Tables = append(schema.Tables,
		TableSchema{Name: "TestTable", Columns: []ColumnSchema{
			{Name: "key", Kind: googlesql.TypeKindTypeInt32},
			{Name: "TestEnum", TypeName: enumLabel},
			{Name: "KitchenSink", TypeName: protoLabel},
		}},
		TableSchema{Name: "EnumTable", Columns: []ColumnSchema{
			{Name: "key", Kind: googlesql.TypeKindTypeInt32},
			{Name: "TestEnum", TypeName: enumLabel},
		}},
	)
	return nil
}

type sampleColumn struct {
	name string
	typ  googlesql.Googlesql_TypeNode
}

func addSampleProtoTable(cat *googlesql.SimpleCatalog, name string, cols []sampleColumn) error {
	tbl, err := googlesql.NewSimpleTable(name, -1)
	if err != nil {
		return fmt.Errorf("new simple table %q: %w", name, err)
	}
	for _, c := range cols {
		col, err := googlesql.NewSimpleColumn(name, c.name, c.typ, false, true)
		if err != nil {
			return fmt.Errorf("new column %s.%s: %w", name, c.name, err)
		}
		if err := tbl.AddColumn(col); err != nil {
			return fmt.Errorf("add column %s.%s: %w", name, c.name, err)
		}
	}
	if err := cat.AddOwnedTable(tbl); err != nil {
		return fmt.Errorf("add table %q: %w", name, err)
	}
	return nil
}
