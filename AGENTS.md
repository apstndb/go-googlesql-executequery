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

Available analyze-only: `none`, `sample`, `tpch`.

Adding a catalog ⇒ hand-port the schema into `catalog/` (no row
data needed; data is unused without an evaluator). Add a unit test
asserting each registered table resolves through `analyze`.

`tpch_graph` returns `ErrUnsupportedCatalog`.

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

## Module path note

The Go module path is `github.com/apstndb/go-googlesql-executequery`.
The working directory was originally created as `…-executesql`
(typo); the user will rename to `…-executequery` after implementation
completes. Implementation does not depend on the directory name; only
the module path matters.
