package catalog

import googlesql "github.com/goccy/go-googlesql"

// sampleSchema is a curated subset of upstream's
// googlesql/testdata/sample_catalog_impl.cc::LoadTables(), chosen to
// cover the tables that the bulk of the analyzer tests reference
// without depending on proto types, enum types, or value-table
// metadata that this Go port has not yet wired through.
//
// In particular this Go port does not yet include:
//   - tables whose columns use proto / enum / struct types
//     (TestTable, EnumTable, ZZZ_AmbiguousHasTestTable, ...): need
//     Go-side proto-descriptor support before they can be modelled.
//   - LazySimpleTable / SimpleTableWithReadTimeIgnored variants:
//     differ only in evaluator-side behaviour, which this port does
//     not exercise.
//   - SetContents row data: this Go port has no evaluator that
//     would consume rows.
//
// The included subset is sufficient for parsing and analyzing the
// most common analyzer-test queries (KeyValue, MultipleColumns,
// TwoIntegers, etc.).
func sampleSchema() *Schema {
	i := googlesql.TypeKindTypeInt64
	str := googlesql.TypeKindTypeString
	ts := googlesql.TypeKindTypeTimestamp
	return &Schema{
		Name: "sample",
		Tables: []TableSchema{
			{Name: "Value", Columns: []ColumnSchema{
				{Name: "Value", Kind: i},
				{Name: "Value_1", Kind: i},
			}},
			{Name: "KeyValue", Columns: []ColumnSchema{
				{Name: "Key", Kind: i},
				{Name: "Value", Kind: str},
			}},
			{Name: "KeyValueLazy", Columns: []ColumnSchema{
				{Name: "Key", Kind: i},
				{Name: "Value", Kind: str},
			}},
			{Name: "KeyValueFindOnly", Columns: []ColumnSchema{
				{Name: "Key", Kind: i},
				{Name: "Value", Kind: str},
			}},
			{Name: "KeyValue2", Columns: []ColumnSchema{
				{Name: "Key", Kind: i},
				{Name: "Value2", Kind: str},
			}},
			{Name: "AnotherKeyValue", Columns: []ColumnSchema{
				{Name: "Key", Kind: i},
				{Name: "value", Kind: str},
			}},
			{Name: "KeyValueWithPrimaryKey", PrimaryKey: []string{"Key"}, Columns: []ColumnSchema{
				{Name: "Key", Kind: i},
				{Name: "Value", Kind: str},
			}},
			{Name: "MultipleColumns", Columns: []ColumnSchema{
				{Name: "int_a", Kind: i},
				{Name: "string_a", Kind: str},
				{Name: "int_b", Kind: i},
				{Name: "string_b", Kind: str},
				{Name: "int_c", Kind: i},
				{Name: "int_d", Kind: i},
			}},
			{Name: "TableWithTimestampColumn", Columns: []ColumnSchema{
				{Name: "ts_col", Kind: ts},
			}},
			{Name: "TableWithMixedCaseColumn", Columns: []ColumnSchema{
				{Name: "ts_col", Kind: ts},
				{Name: "MiXeD_cAsE", Kind: str},
			}},
			{Name: "abTable", Columns: []ColumnSchema{
				{Name: "a", Kind: i},
				{Name: "b", Kind: str},
			}},
			{Name: "bcTable", Columns: []ColumnSchema{
				{Name: "b", Kind: i},
				{Name: "c", Kind: str},
			}},
			{Name: "TwoIntegers", PrimaryKey: []string{"key"}, Columns: []ColumnSchema{
				{Name: "key", Kind: i},
				{Name: "value", Kind: i},
			}},
			{Name: "FourIntegers", PrimaryKey: []string{"key1", "key2"}, Columns: []ColumnSchema{
				{Name: "key1", Kind: i},
				{Name: "value1", Kind: i},
				{Name: "key2", Kind: i},
				{Name: "value2", Kind: i},
			}},
			{Name: "NoColumns"},
		},
	}
}
