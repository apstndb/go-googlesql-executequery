# AGENTS.md

Project-specific guidance for any AI agent collaborating on this
repository. The CLAUDE.md symlink at the root points to this file.

## Language policy

English only. This applies to:

- All code, comments, and documentation in this repository.
- Commit messages.
- GitHub issues and pull requests filed against this repository.

Discussion threads with the user can be in any language; only artefacts
that land in this repo are constrained.

## Build and test

```sh
mise install                    # installs Go and golangci-lint at pinned versions
git submodule update --init     # populates third_party/googlesql
mise run ci                     # lint + test + build
mise run test:corpus            # corpus golden tests (separate from ci)
```

Do not introduce a `Makefile`, `Taskfile.yml`, or shell scripts that
duplicate `mise.toml`. Add new automation as `mise.toml` tasks.

All dev-tool versions live in `mise.toml`. Pin new tools there.

## Versioning

Releases use `v<go-googlesql-Major.Minor>.<YYYYMMDD>` with date
carry-over on collision. See `RELEASING.md`.

## Submodule policy

`third_party/googlesql/` is a git submodule of `google/googlesql`,
sparse-checked-out to a small subset of upstream files. It exists
because this project is an Apache-2.0 derivative of `execute_query`,
and the submodule supplies:

- the `execute_query.md` user-facing documentation,
- the C++ source we hand-port catalog schemas from
  (`googlesql/testdata/sample_catalog.cc`),
- the TPCH `describe.txt` schema definitions,
- the SQL example corpus used by `test:corpus`,
- the verbatim `ABSL_FLAG(...)` help strings used by the CLI.

Do not modify upstream files in-place. Bump the submodule by updating
its commit pin and re-running `mise run ci`.

## Modes (current)

Supported: `parse`, `unparse`, `analyze` (alias `resolve`).

Unsupported (return `ErrUnsupportedMode`):

- `unanalyze` / `sql_builder` — `goccy/go-googlesql` does not export
  `SQLBuilder`. Unblocked when upstream exposes a Resolved → SQL
  function.
- `explain` — `goccy/go-googlesql` does not export the reference
  evaluator. Unblocked when upstream ships `PreparedQuery` /
  `PreparedStatement`.
- `execute` — same as `explain`.

When adding new modes, mirror the C++ `ExecuteQuery` dispatcher in
`executequery.go` and add per-mode tests.

## Flag policy

Every C++ `execute_query` flag is either:

- **fully wired** to `AnalyzerOptions` / `LanguageOptions`; or
- **defined and rejected** when set to a non-default value.

Each unsupported flag has a Go source-code comment block at its
declaration (in `cmd/execute_query/flags.go` or `unsupported.go`)
documenting:

1. what the flag does in upstream,
2. why it cannot be honoured today (the specific missing API), and
3. what upstream change unblocks it.

The same prose is exposed as the runtime error suffix and as the
`--help` annotation, via a per-flag `unsupportedReason*` constant.

CLI help text is **verbatim** from the upstream `ABSL_FLAG(...)`
declarations. Do not paraphrase. When upstream changes a help
string, update the constant and the submodule pin together.

## Catalogs

Available analyze-only: `none`, `sample`, `tpch`, `tpch_graph`.

`tpch_graph` adds the join pseudo-columns from upstream's
`tpch_catalog.cc` (e.g. `Customer.Orders : MULTIROW<Orders>`). Walking
those pseudo-columns requires
`--enabled_language_features=ALL_MINUS_DEV,+FEATURE_ROW_TYPE` (the
`--catalog` help text says so). The
`Column::JoinColumnAttributes` upstream attaches to each pseudo-column
is **not** set — `goccy.OptionalJoinColumnAttributes` has no exported
constructor or `SimpleColumn` setter — so behaviour that relies on
those attributes (e.g. upstream's join-flattening rewrite) will
diverge.

Adding a catalog ⇒ hand-port the schema into `catalog/` (no row
data needed; data is unused without an evaluator). Add a unit test
asserting each registered table resolves through `analyze`.

For catalogs that need types beyond the scalar `TypeKind` set —
proto messages, enums, structs — register them through the
`PostBuild` hook on `buildSimple`. The hook receives the freshly
constructed `*SimpleCatalog`, the `*Schema` (mutate it to mirror
extra tables into DESCRIBE output), the per-table `SimpleTable`
handle map, and the `*TypeFactory`. From the hook you can:
`SetDescriptorPool` to attach a `*googlesql.DescriptorPool` populated
via `NewFileDescriptorProto.ParseFromString` + `BuildFile`; resolve
`*Descriptor`/`*EnumDescriptor` via `FindMessageTypeByName`/
`FindEnumTypeByName`; build types via `TypeFactory.MakeProtoType`/
`MakeEnumType`/`MakeStructType`; and `AddOwnedTable` directly for
tables that don't fit the Schema → buildTable flow.
`catalog/sample_proto.go` is the reference implementation —
hand-builds a minimal `descriptorpb.FileDescriptorProto` (no protoc
dep), so `google.golang.org/protobuf` is the only Go-side dependency
this path adds.

## Cache directory

Always resolve through `cache.Default()` or honour `--cache_dir`.
Never write under `/tmp` or the current working directory.

The cache subdir is keyed by the linked `goccy/go-googlesql` module
version (read from `runtime/debug.ReadBuildInfo`) — the wasm ABI
binds to upstream, not to this wrapper.

## Investigation dependencies

Clone exploratory clones into `.tmp/` (gitignored). Do not commit
them. The submodule under `third_party/googlesql` is the only
pinned-by-this-repo dependency.

## Workaround comment convention

Whenever code works around a missing or buggy `goccy/go-googlesql`
API, the comment at the workaround site must spell out **what the
natural code would be** if the API behaved as expected, so future
readers can grep and so the workaround is easy to revert when
upstream lands a fix.

Use this shape (a `Workaround:` block, with a `Natural code:` line
showing the call we wish we could make and a follow-on explaining
what we do instead):

```go
// Workaround for goccy/go-googlesql v0.2.1: <one-line bug summary>.
//
// Natural code:
//   <2-3 lines of the obvious goccy call we would make>
//
// Instead, <what this code does and why> Unblocked when:
// <upstream change that lets us delete the workaround>.
```

The error / `--help` strings under `unsupported.go` follow a
parallel "what / why / unblocked when" shape; reuse the same
phrasing so users and code readers see consistent reasoning.

## goccy/go-googlesql gotchas

Quirks of `goccy/go-googlesql` v0.2.1 that are not obvious from the
public API and have already cost real debugging time:

- `ParserOptions.SetLanguageOptions(lo)` *moves-from* its argument on
  the wasm side. The caller's `*LanguageOptions` survives as a Go
  handle but now points at a default-constructed instance with no
  enabled features. Hand the parser its own freshly-built copy
  (`executequery.go` builds a separate `parserLO` exactly for this);
  never share one `*LanguageOptions` between `ParserOptions` and
  `AnalyzerOptions`.
- `SimpleCatalog.AddOwnedTable(table)` calls `clearPtrAny(table)` after
  the wasm trip, so any retained `*SimpleTable` handle becomes null.
  If you need to mutate tables after they are part of a catalog,
  populate them *before* calling `AddOwnedTable` (see the `postBuild`
  hook on `catalog.buildSimple`), or switch to `AddTable` and own the
  table from Go.
- `goccy.OptionalJoinColumnAttributes` is exported but has no
  constructor or `SimpleColumn` setter. There is currently no way to
  attach `Column::JoinColumnAttributes` from Go.
