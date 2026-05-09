package catalog

import googlesql "github.com/goccy/go-googlesql"

// tpchSchema is the standard TPC-H benchmark schema as documented in
// third_party/googlesql/googlesql/examples/tpch/describe.txt.
//
// The Go port stores no row data; the catalog is usable only for
// parse / analyze. (Upstream's selectable `tpch` catalog ships
// 1 MB of CSV data plus a CSV reader; the reference evaluator that
// would consume it is not exposed by goccy/go-googlesql.)
//
// Column types and primary keys match upstream describe.txt exactly.
func tpchSchema() *Schema {
	u := googlesql.TypeKindTypeUint64
	i := googlesql.TypeKindTypeInt64
	d := googlesql.TypeKindTypeDouble
	s := googlesql.TypeKindTypeString
	dt := googlesql.TypeKindTypeDate
	return &Schema{
		Name: "tpch",
		Tables: []TableSchema{
			{Name: "Customer", PrimaryKey: []string{"C_CUSTKEY"}, Columns: []ColumnSchema{
				{Name: "C_CUSTKEY", Kind: u},
				{Name: "C_NAME", Kind: s},
				{Name: "C_ADDRESS", Kind: s},
				{Name: "C_NATIONKEY", Kind: u},
				{Name: "C_PHONE", Kind: s},
				{Name: "C_ACCTBAL", Kind: d},
				{Name: "C_MKTSEGMENT", Kind: s},
				{Name: "C_COMMENT", Kind: s},
			}},
			{Name: "LineItem", PrimaryKey: []string{"L_ORDERKEY", "L_LINENUMBER"}, Columns: []ColumnSchema{
				{Name: "L_ORDERKEY", Kind: u},
				{Name: "L_PARTKEY", Kind: u},
				{Name: "L_SUPPKEY", Kind: u},
				{Name: "L_LINENUMBER", Kind: u},
				{Name: "L_QUANTITY", Kind: d},
				{Name: "L_EXTENDEDPRICE", Kind: d},
				{Name: "L_DISCOUNT", Kind: d},
				{Name: "L_TAX", Kind: d},
				{Name: "L_RETURNFLAG", Kind: s},
				{Name: "L_LINESTATUS", Kind: s},
				{Name: "L_SHIPDATE", Kind: dt},
				{Name: "L_COMMITDATE", Kind: dt},
				{Name: "L_RECEIPTDATE", Kind: dt},
				{Name: "L_SHIPINSTRUCT", Kind: s},
				{Name: "L_SHIPMODE", Kind: s},
				{Name: "L_COMMENT", Kind: s},
			}},
			{Name: "Nation", PrimaryKey: []string{"N_NATIONKEY"}, Columns: []ColumnSchema{
				{Name: "N_NATIONKEY", Kind: u},
				{Name: "N_NAME", Kind: s},
				{Name: "N_REGIONKEY", Kind: u},
				{Name: "N_COMMENT", Kind: s},
			}},
			{Name: "Orders", PrimaryKey: []string{"O_ORDERKEY"}, Columns: []ColumnSchema{
				{Name: "O_ORDERKEY", Kind: u},
				{Name: "O_CUSTKEY", Kind: u},
				{Name: "O_ORDERSTATUS", Kind: s},
				{Name: "O_TOTALPRICE", Kind: d},
				{Name: "O_ORDERDATE", Kind: dt},
				{Name: "O_ORDERPRIORITY", Kind: s},
				{Name: "O_CLERK", Kind: s},
				{Name: "O_SHIPPRIORITY", Kind: i},
				{Name: "O_COMMENT", Kind: s},
			}},
			{Name: "Part", PrimaryKey: []string{"P_PARTKEY"}, Columns: []ColumnSchema{
				{Name: "P_PARTKEY", Kind: u},
				{Name: "P_NAME", Kind: s},
				{Name: "P_MFGR", Kind: s},
				{Name: "P_BRAND", Kind: s},
				{Name: "P_TYPE", Kind: s},
				{Name: "P_SIZE", Kind: i},
				{Name: "P_CONTAINER", Kind: s},
				{Name: "P_RETAILPRICE", Kind: d},
				{Name: "P_COMMENT", Kind: s},
			}},
			{Name: "PartSupp", PrimaryKey: []string{"PS_PARTKEY", "PS_SUPPKEY"}, Columns: []ColumnSchema{
				{Name: "PS_PARTKEY", Kind: u},
				{Name: "PS_SUPPKEY", Kind: u},
				{Name: "PS_AVAILQTY", Kind: i},
				{Name: "PS_SUPPLYCOST", Kind: d},
				{Name: "PS_COMMENT", Kind: s},
			}},
			{Name: "Region", PrimaryKey: []string{"R_REGIONKEY"}, Columns: []ColumnSchema{
				{Name: "R_REGIONKEY", Kind: u},
				{Name: "R_NAME", Kind: s},
				{Name: "R_COMMENT", Kind: s},
			}},
			{Name: "Supplier", PrimaryKey: []string{"S_SUPPKEY"}, Columns: []ColumnSchema{
				{Name: "S_SUPPKEY", Kind: u},
				{Name: "S_NAME", Kind: s},
				{Name: "S_ADDRESS", Kind: s},
				{Name: "S_NATIONKEY", Kind: u},
				{Name: "S_PHONE", Kind: s},
				{Name: "S_ACCTBAL", Kind: d},
				{Name: "S_COMMENT", Kind: s},
			}},
		},
	}
}
