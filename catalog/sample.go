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
				{"Value", i},
				{"Value_1", i},
			}},
			{Name: "KeyValue", Columns: []ColumnSchema{
				{"Key", i},
				{"Value", str},
			}},
			{Name: "KeyValueLazy", Columns: []ColumnSchema{
				{"Key", i},
				{"Value", str},
			}},
			{Name: "KeyValueFindOnly", Columns: []ColumnSchema{
				{"Key", i},
				{"Value", str},
			}},
			{Name: "KeyValue2", Columns: []ColumnSchema{
				{"Key", i},
				{"Value2", str},
			}},
			{Name: "AnotherKeyValue", Columns: []ColumnSchema{
				{"Key", i},
				{"value", str},
			}},
			{Name: "KeyValueWithPrimaryKey", PrimaryKey: []string{"Key"}, Columns: []ColumnSchema{
				{"Key", i},
				{"Value", str},
			}},
			{Name: "MultipleColumns", Columns: []ColumnSchema{
				{"int_a", i},
				{"string_a", str},
				{"int_b", i},
				{"string_b", str},
				{"int_c", i},
				{"int_d", i},
			}},
			{Name: "TableWithTimestampColumn", Columns: []ColumnSchema{
				{"ts_col", ts},
			}},
			{Name: "TableWithMixedCaseColumn", Columns: []ColumnSchema{
				{"ts_col", ts},
				{"MiXeD_cAsE", str},
			}},
			{Name: "abTable", Columns: []ColumnSchema{
				{"a", i},
				{"b", str},
			}},
			{Name: "bcTable", Columns: []ColumnSchema{
				{"b", i},
				{"c", str},
			}},
			{Name: "TwoIntegers", PrimaryKey: []string{"key"}, Columns: []ColumnSchema{
				{"key", i},
				{"value", i},
			}},
			{Name: "FourIntegers", PrimaryKey: []string{"key1", "key2"}, Columns: []ColumnSchema{
				{"key1", i},
				{"value1", i},
				{"key2", i},
				{"value2", i},
			}},
			{Name: "NoColumns"},
		},
	}
}
