package executequery

import (
	"errors"
	"fmt"
)

// Sentinel errors. Wrapped at use-site with a per-feature reason
// constant so callers can errors.Is the category and end users see
// the upstream-blocking detail in the message.
var (
	ErrUnsupportedMode    = errors.New("mode not supported by go-googlesql")
	ErrUnsupportedSQLMode = errors.New("sql_mode not supported by go-googlesql")
	ErrUnsupportedCatalog = errors.New("catalog not supported by go-googlesql")
	ErrUnsupportedFlag    = errors.New("flag not supported by go-googlesql")
)

// Per-feature reason strings.
//
// Each constant carries a comment block explaining:
//   1. what the upstream feature does,
//   2. why it cannot be honoured today (the specific missing API), and
//   3. what change unblocks it.
//
// The same string is used as the runtime error suffix and as the
// `--help` annotation for unsupported flags, so users and code
// readers see the identical reasoning.

// ReasonModeUnanalyze: upstream's `unanalyze` (alias `sql_builder`)
// converts a Resolved AST back to SQL via SQLBuilder, with optional
// dialect selection through `--target_syntax`.
//
// Why unsupported: go-googlesql does not export SQLBuilder or
// any other Resolved → SQL function.
//
// Unblocked when: go-googlesql exposes a Resolved → SQL builder
// (tracked at: TBD upstream issue once filed).
const ReasonModeUnanalyze = "" +
	"the unanalyze mode (Resolved AST → SQL via SQLBuilder) is not exposed by go-googlesql"

// ReasonModeExplain: upstream's `explain` mode prints the reference
// implementation's evaluator query plan for a resolved statement.
//
// Why unsupported: go-googlesql does not export the reference
// evaluator (no PreparedQuery / PreparedStatement / Evaluator
// constructors are public).
//
// Unblocked when: go-googlesql exposes the reference evaluator
// (or at least its plan-string output).
const ReasonModeExplain = "" +
	"the explain mode requires the reference evaluator, which go-googlesql does not expose"

// ReasonModeExecute: upstream's `execute` mode runs the statement
// through the reference implementation's evaluator and renders rows.
//
// Why unsupported: same as ReasonModeExplain — the reference
// evaluator is not exposed by go-googlesql.
//
// Unblocked when: go-googlesql exposes PreparedQuery /
// PreparedStatement (or equivalent execution primitives).
const ReasonModeExecute = "" +
	"the execute mode requires the reference evaluator, which go-googlesql does not expose"

// ReasonFlagTargetSyntax: upstream's `--target_syntax` selects the
// SQL dialect produced by the unanalyze mode (`standard` or `pipe`).
//
// Why unsupported: only meaningful for unanalyze, which is itself
// unsupported (see ReasonModeUnanalyze).
const ReasonFlagTargetSyntax = "" +
	"--target_syntax only affects the unanalyze mode, which is not supported"

// ReasonFlagUseBoxGlyphs: upstream's `--use_box_glyphs` toggles
// Unicode box characters in resolved-AST and result-table output.
//
// Why unsupported: the result-table renderer is execute-only, and
// the resolved-AST output we emit (via DebugString) is upstream's
// own ASCII-only rendering — there is no glyph choice to make.
const ReasonFlagUseBoxGlyphs = "" +
	"--use_box_glyphs only affects the execute-mode result table, which is not supported"

// ReasonFlagOutputMode: upstream's `--output_mode` controls
// execute-mode result rendering (box / json / textproto).
//
// Why unsupported: the parse / analyze modes use upstream's own
// DebugString text format. Emitting AST or Resolved nodes as
// textproto / json would require Serialize(*Proto) on the AST and
// Resolved node types, which go-googlesql does not expose.
//
// Note on the proto path: the C++ side inside the wasm module was
// built with protoc-gen-cpp, so every *Proto type there already
// implements google::protobuf::Message and can exchange wire-format
// bytes with a protoc-gen-go struct of the same definition.  If
// go-googlesql exposed Serialize() on AST / Resolved handles, the
// resulting *Proto wire bytes could be fed straight into
// google.golang.org/protobuf/protojson.Marshal or
// prototext.Marshal to produce JSON / textproto output with no
// hand-written visitor.  The only missing link is the Handle ->
// Proto conversion method on the wasm side.
//
// Unblocked when: go-googlesql exposes Serialize on AST and
// Resolved node types (or the project commits to an in-process
// visitor implementation, which is significant work).
const ReasonFlagOutputMode = "" +
	"--output_mode is execute-only; AST/Resolved Serialize is not exposed by go-googlesql"

// ReasonFlagTableSpec: upstream's `--table_spec` registers tables
// from CSV / binproto / textproto files for execute mode to query.
//
// Why unsupported: row data is only meaningful for execute mode,
// which is unsupported.
const ReasonFlagTableSpec = "" +
	"--table_spec attaches data to tables for execute mode, which is not supported"

// ReasonFlagDescriptorPool: upstream's `--descriptor_pool` selects
// the proto descriptor pool used to resolve proto types
// (`generated` = the C++ generated pool; `none` = no proto types).
//
// Why unsupported: the C++ generated pool relies on protoc-compiled
// types linked into the upstream binary. Crossing the wasm boundary
// requires extra plumbing in go-googlesql to register Go
// proto descriptors with the wasm runtime.
//
// Unblocked when: go-googlesql exposes a way to register a Go
// (protoreflect / google.golang.org/protobuf) descriptor pool with
// the wasm runtime.
const ReasonFlagDescriptorPool = "" +
	"--descriptor_pool requires wasm-boundary descriptor-pool plumbing, which go-googlesql does not expose"

// ReasonFlagEvaluatorMaxValue: upstream's
// `--evaluator_max_value_byte_size` caps per-Value memory in execute
// mode.
//
// Why unsupported: evaluator-only flag.
const ReasonFlagEvaluatorMaxValue = "" +
	"--evaluator_max_value_byte_size only affects the execute mode, which is not supported"

// ReasonFlagEvaluatorMaxIntermediate: upstream's
// `--evaluator_max_intermediate_byte_size` caps accumulated row
// memory in execute mode.
//
// Why unsupported: evaluator-only flag.
const ReasonFlagEvaluatorMaxIntermediate = "" +
	"--evaluator_max_intermediate_byte_size only affects the execute mode, which is not supported"

// ReasonFlagEvaluatorScramble: upstream's
// `--evaluator_scramble_undefined_orderings` shuffles unordered
// intermediate results in execute mode (a determinism-canary).
//
// Why unsupported: evaluator-only flag.
const ReasonFlagEvaluatorScramble = "" +
	"--evaluator_scramble_undefined_orderings only affects the execute mode, which is not supported"

// ReasonFlagMaxStatementsToExecute: upstream's
// `--max_statements_to_execute` caps statements executed in script
// mode.
//
// Why unsupported: script execution requires the reference evaluator.
// Script parsing and per-statement analysis are supported, but no
// execution count is meaningful.
const ReasonFlagMaxStatementsToExecute = "" +
	"--max_statements_to_execute caps script execution, which is not supported"

// ReasonFlagImportPath: upstream's `--import_path` adds directories
// to search for IMPORT MODULE.
//
// Why unsupported: go-googlesql does not expose ModuleFactory,
// so resolution of IMPORT MODULE statements cannot be wired through.
//
// Unblocked when: go-googlesql exposes ModuleFactory.
const ReasonFlagImportPath = "" +
	"--import_path requires ModuleFactory, which go-googlesql does not expose"

// ReasonFlagWeb: upstream's `--web` runs a local HTTP server with a
// query-and-execute UI.
//
// Why unsupported: without execute mode the UI loses its primary
// utility; not currently planned for this Go port.
const ReasonFlagWeb = "" +
	"--web is not implemented in this Go port (and would have limited value without execute mode)"

// ReasonFlagPort: upstream's `--port` selects the local HTTP port
// for `--web`.
//
// Why unsupported: tied to --web.
const ReasonFlagPort = "" +
	"--port only affects --web, which is not implemented"

// Proto-binding note
//
// Many of the gaps above are rooted in the same architectural fact:
// go-googlesql's *Proto types are wrappers around C++
// google::protobuf::Message objects (generated by protoc-gen-cpp), while
// idiomatic Go code uses google.golang.org/protobuf types (generated by
// protoc-gen-go).  Because both sides speak the same protobuf wire
// format, any *Proto value can already be converted to/from a standard
// Go struct by calling ParseFromString / SerializeToString on the
// go-googlesql side and proto.Marshal / proto.Unmarshal on the Go side.
//
// What is missing is the bridge *from* Handle types *to* those Proto
// wrappers.  For example:
//
//   - AnalyzerOptions has no Deserialize(*AnalyzerOptionsProto) method,
//     so we cannot load analyzer settings from a JSON/YAML config file.
//   - AST nodes and Resolved nodes have no Serialize(*Proto) methods,
//     so we cannot emit structured AST output (JSON / textproto).
//   - Value (evaluator results) has no Serialize(*ValueProto) method,
//     so execute-mode result tables cannot be emitted as JSON/textproto.
//
// If go-googlesql adds those Handle->Proto / Proto->Handle conversion
// methods, the missing features could be implemented almost entirely
// with standard protoc-gen-go types plus the existing wire-format
// shunt, without new wasm-side code.
//
// The one place where this already works today is catalog/sample_proto.go:
// descriptorpb.FileDescriptorProto (protoc-gen-go) -> wire bytes ->
// googlesql.FileDescriptorProto (protoc-gen-wasmify-go) ->
// DescriptorPool.BuildFile.  That pattern is the model for what would
// unlock the larger features above.

// modeUnsupportedf builds an error describing an unsupported Mode.
func modeUnsupportedf(name, reason string) error {
	return fmt.Errorf("%w: %q: %s", ErrUnsupportedMode, name, reason)
}

func sqlModeUnsupportedf(name, reason string) error {
	return fmt.Errorf("%w: %q: %s", ErrUnsupportedSQLMode, name, reason)
}

func catalogUnsupportedf(name, reason string) error {
	return fmt.Errorf("%w: %q: %s", ErrUnsupportedCatalog, name, reason)
}

// FlagUnsupportedError builds an error describing an unsupported
// flag set to a non-default value. Exported so the CLI layer can
// produce identical messages for every rejected flag.
func FlagUnsupportedError(name, reason string) error {
	return fmt.Errorf("%w: --%s: %s", ErrUnsupportedFlag, name, reason)
}
