package main

// Help text for every flag mirrors upstream's ABSL_FLAG declarations
// in third_party/googlesql/googlesql/tools/execute_query/execute_query_tool.cc
// verbatim. Each unsupported flag's help is suffixed with a
// "(NOT SUPPORTED ...)" annotation drawn from
// executequery.ReasonFlag* so users see the upstream-side blocker
// inline in --help.

// helpProductMode is verbatim from upstream's --product_mode flag.
const helpProductMode = "" +
	"The product_mode to use in language options. Note, language_features" +
	" is an orthongal way to configure language options." +
	"\nValid values are:" +
	"\n     'internal': supports protos, DOUBLE, signed ints, etc. " +
	"\n     'external': mode used in Cloud engines"

// helpMode is verbatim from upstream's --mode flag.
const helpMode = "" +
	"The comma-separated tool modes to use. Valid values are:" +
	"\n     'parse'   parse the parser AST" +
	"\n     'unparse'  parse, then dump as sql" +
	"\n     'analyze'  print the resolved AST" +
	"\n     'unanalyze'  analyze, then dump as sql" +
	"\n     'explain'  print the evaluator query plan" +
	"\n     'execute'  actually run the query and print the result. (not" +
	"                  all functionality is supported)."

// helpCatalog mirrors upstream's --catalog (the upstream string is
// "The base catalog to use ..." plus runtime-generated catalog
// descriptions; we hard-code the descriptions to avoid depending on
// upstream's StrCat helper).
const helpCatalog = "" +
	"The base catalog to use for looking up tables, etc.\nChoices:" +
	"\n  'none' - empty catalog" +
	"\n  'sample' - the analyzer-test sample catalog" +
	"\n  'tpch' - the TPC-H benchmark schema" +
	"\n  'tpch_graph' - TPC-H plus a property graph (NOT SUPPORTED in this Go port)"

const helpImportPath = "The comma-separated list of directories to search for modules."

const helpEnabledASTRewrites = "" +
	"The AST Rewrites to enable in the analyzer, format is:" +
	"\n   <BASE>[,+<ADDED_OPTION>][,-<REMOVED_OPTION>]..." +
	"\n Where BASE is one of:" +
	"\n   'NONE': The empty set" +
	"\n   'ALL': All possible rewrites, including those in development. Not recommended, in-development rewrites may produce incorrect results" +
	"\n   'ALL_MINUS_DEV': (Default) All rewrites except those in development" +
	"\n   'DEFAULTS': All ResolvedASTRewrite's with 'default_enabled' set. Not recommended, in-development rewrites may produce incorrect results" +
	"\n   'DEFAULTS_MINUS_DEV': All rewrites with 'default_enabled' set, except those in development" +
	"\n" +
	"\n Enum values must be listed with 'REWRITE_' stripped" +
	"\n Example:" +
	"\n    --enabled_ast_rewrites='DEFAULTS,-FLATTEN,+ANONYMIZATION'" +
	"\n Will enable all the default options plus ANONYMIZATION, but excluding flatten"

const helpFoldLiteralCast = "Set the fold_literal_cast option in AnalyzerOptions"

const helpEnabledLanguageFeatures = "" +
	"The set of LanguageFeatures to enable in the analyzer, format is:" +
	"\n   <BASE>[,+<ADDED_OPTION>][,-<REMOVED_OPTION>]..." +
	"\n Where BASE is one of NONE | ALL | ALL_MINUS_DEV | DEFAULTS | DEFAULTS_MINUS_DEV" +
	"\n Feature names use the upstream FEATURE_* spelling and may be supplied with or" +
	" without underscores."

const helpParameters = "" +
	"Query parameters as a comma-separated list of name=value pairs. " +
	"Types are inferred from the value: integers → INT64, floats → DOUBLE, " +
	"'string' / \"string\" → STRING, TRUE/FALSE → BOOL."

const helpStrictNameResolutionMode = "Sets LanguageOptions::strict_resolution_mode."

const helpEvaluatorScrambleUndefinedOrderings = "" +
	"When true, shuffle the order of rows in intermediate results that " +
	"are unordered."

const helpPruneUnusedColumns = "Sets AnalyzerOptions::prune_unused_columns."

const helpParseLocationRecordType = "" +
	"Value for AnalyzerOptions::parse_location_record_type." +
	"\nValid values are:" +
	"\n  'NONE': Do not record locations" +
	"\n  'FULL_NODE_SCOPE': Locations cover full range of the related node" +
	"\n  'CODE_SEARCH': Locations cover related object name relevant for code search"

const helpTableSpec = "" +
	"The table spec to use for building the GoogleSQL Catalog. This is a " +
	"comma-delimited list of strings of the form <table_name>=<spec>, " +
	"where <spec> is of the form:" +
	"\n    binproto:<proto>:<path> - binary proto file that is represented by a value table" +
	"\n    textproto:<proto>:<path> - text proto file that is represented by a value table" +
	"\n    csv:<path> - csv file that is represented by a table whose string-typed column names are determined from the header row."

const helpDescriptorPool = "" +
	"The descriptor pool to use while resolving the query. This can be:" +
	"\n    'generated' - the generated pool of protos compiled into this binary" +
	"\n    'none'      - no protos are included (but syntax is still supported"

const helpOutputMode = "" +
	"Format to use for query results. Available choices:" +
	"\nbox - Tabular format for human consumption" +
	"\njson - JSON serialization" +
	"\ntextproto - Protocol buffer text format"

const helpSQLMode = "" +
	"How to interpret the input sql. Available choices:" +
	"\nquery" +
	"\nexpression" +
	"\nscript"

const helpTargetSyntax = "" +
	"The syntax to use when generating SQL from the resolved AST. " +
	"Available choices:" +
	"\nstandard - The standard syntax with nested subqueries" +
	"\npipe - The pipe syntax with flattened subqueries"

const helpEvaluatorMaxValueByteSize = `Limit on the maximum number of in-memory bytes used by an individual Value
  that is constructed during evaluation. This bound applies to all Value
  types, including variable-sized types like STRING, BYTES, ARRAY, and
  STRUCT. Exceeding this limit results in an error. See the implementation of
  Value::physical_byte_size for more details.`

const helpEvaluatorMaxIntermediateByteSize = `The limit on the maximum number of in-memory bytes that can be used for
  storing accumulated rows (e.g., during an ORDER BY query). Exceeding this
  limit results in an error.`

const helpMaxStatementsToExecute = "" +
	"The limit on number of statements allowed for execution. Post this " +
	"limit, script is considered to have infinite loop and returned error."

const helpUseBoxGlyphs = "" +
	"Use Unicode box glyphs instead of ASCII characters for the resolved " +
	"AST and the result tables output."

// Go-port-specific flag help (not from upstream).

const helpCacheDir = "" +
	"Directory used as wazero's compilation cache for the embedded " +
	"GoogleSQL wasm module. When unset, defaults to a per-user, " +
	"per-go-googlesql-version directory under os.UserCacheDir()."

const helpNoCache = "" +
	"Disable the on-disk wazero compilation cache. Forces interpreter " +
	"mode (slower at runtime; no native-code cache to manage)."

const helpCompilationMode = "" +
	"wazero compilation mode for the embedded GoogleSQL wasm module: " +
	"'compiler' (default; native code) or 'interpreter' (slower; less RAM)."

const helpSQLFile = "" +
	"Read SQL from the given file instead of the positional argument. " +
	"A leading '@' on the positional argument has the same effect; '-' " +
	"reads from stdin."

const helpWeb = "Run a local webserver to execute queries."

const helpPort = "Port to run the local webserver on."
