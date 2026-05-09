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
// Column types match upstream describe.txt exactly. Primary keys
// from describe.txt are not yet captured here because
// SimpleTable.SetPrimaryKey requires column-index plumbing we do
// not need for analyze; add when DESCRIBE wants to surface them.
func tpchSchema() *Schema {
	u := googlesql.TypeKindTypeUint64
	i := googlesql.TypeKindTypeInt64
	d := googlesql.TypeKindTypeDouble
	s := googlesql.TypeKindTypeString
	dt := googlesql.TypeKindTypeDate
	return &Schema{
		Name: "tpch",
		Tables: []TableSchema{
			{Name: "Customer", Columns: []ColumnSchema{
				{"C_CUSTKEY", u},
				{"C_NAME", s},
				{"C_ADDRESS", s},
				{"C_NATIONKEY", u},
				{"C_PHONE", s},
				{"C_ACCTBAL", d},
				{"C_MKTSEGMENT", s},
				{"C_COMMENT", s},
			}},
			{Name: "LineItem", Columns: []ColumnSchema{
				{"L_ORDERKEY", u},
				{"L_PARTKEY", u},
				{"L_SUPPKEY", u},
				{"L_LINENUMBER", u},
				{"L_QUANTITY", d},
				{"L_EXTENDEDPRICE", d},
				{"L_DISCOUNT", d},
				{"L_TAX", d},
				{"L_RETURNFLAG", s},
				{"L_LINESTATUS", s},
				{"L_SHIPDATE", dt},
				{"L_COMMITDATE", dt},
				{"L_RECEIPTDATE", dt},
				{"L_SHIPINSTRUCT", s},
				{"L_SHIPMODE", s},
				{"L_COMMENT", s},
			}},
			{Name: "Nation", Columns: []ColumnSchema{
				{"N_NATIONKEY", u},
				{"N_NAME", s},
				{"N_REGIONKEY", u},
				{"N_COMMENT", s},
			}},
			{Name: "Orders", Columns: []ColumnSchema{
				{"O_ORDERKEY", u},
				{"O_CUSTKEY", u},
				{"O_ORDERSTATUS", s},
				{"O_TOTALPRICE", d},
				{"O_ORDERDATE", dt},
				{"O_ORDERPRIORITY", s},
				{"O_CLERK", s},
				{"O_SHIPPRIORITY", i},
				{"O_COMMENT", s},
			}},
			{Name: "Part", Columns: []ColumnSchema{
				{"P_PARTKEY", u},
				{"P_NAME", s},
				{"P_MFGR", s},
				{"P_BRAND", s},
				{"P_TYPE", s},
				{"P_SIZE", i},
				{"P_CONTAINER", s},
				{"P_RETAILPRICE", d},
				{"P_COMMENT", s},
			}},
			{Name: "PartSupp", Columns: []ColumnSchema{
				{"PS_PARTKEY", u},
				{"PS_SUPPKEY", u},
				{"PS_AVAILQTY", i},
				{"PS_SUPPLYCOST", d},
				{"PS_COMMENT", s},
			}},
			{Name: "Region", Columns: []ColumnSchema{
				{"R_REGIONKEY", u},
				{"R_NAME", s},
				{"R_COMMENT", s},
			}},
			{Name: "Supplier", Columns: []ColumnSchema{
				{"S_SUPPKEY", u},
				{"S_NAME", s},
				{"S_ADDRESS", s},
				{"S_NATIONKEY", u},
				{"S_PHONE", s},
				{"S_ACCTBAL", d},
				{"S_COMMENT", s},
			}},
		},
	}
}
