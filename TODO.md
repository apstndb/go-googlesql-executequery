# TODO

Living checklist for follow-up work. Keep entries terse; move detail
into `unsupported.go` reason constants or commit messages.

## Release / housekeeping (post-implementation)

- [ ] Rename working directory `go-googlesql-executesql` → `go-googlesql-executequery`.
      The Go module path is already `…-executequery`; only the dir
      name on disk needs fixing.
- [x] Initial commit and push.
      Confirm there are no embedded secrets or stray cache directories
      first (`.tmp/` is gitignored; double-check `git status`).
- [x] Create the GitHub repo at `apstndb/go-googlesql-executequery` and
      push `main` plus the initial release tag.
- [x] Cut the first release: `v0.2.YYYYMMDD` (go-googlesql v0.2.x line).
      See `RELEASING.md` for the carry-over rule. (Shipped `v0.2.20260509`.)
- [ ] Once CI is green on `main`, double-check the macOS leg covers
      the `cache.Default()` path that XDG-only Linux runners do not.

## Documentation polish

- [ ] Replace each "(tracked at: TBD upstream issue once filed)" in
      `unsupported.go` with a real link, after filing the issues
      against `goccy/go-googlesql`.
- [x] Flesh out `RELEASING.md` with concrete CI-verification steps
      (which workflow to watch, how to retry on flake, etc.).
- [x] Add a usage example block to `cmd/execute_query/main.go`'s
      `usage` string (the upstream `kUsage` only shows the bare
      shape; readers benefit from a `--mode=parse "SELECT 1"` example).

## v0.2+ feature unlocks (gated on upstream)

Each item below is recognised today as a flag/value and rejected with
a structured `ErrUnsupportedFlag` / `ErrUnsupportedMode` /
`ErrUnsupportedCatalog`. The reason strings in `unsupported.go`
spell out the precise upstream change that unblocks each one.

- [ ] `--mode=unanalyze` / `sql_builder` and `--target_syntax`.
      Needs a Resolved-AST → SQL builder exposed by go-googlesql
      (upstream's `SQLBuilder`).
- [ ] `--mode=execute` and `--mode=explain`.
      Needs the reference evaluator (`PreparedQuery`,
      `PreparedStatement`, or equivalent) exposed by go-googlesql.
      Unblocks `--output_mode=box`, `--use_box_glyphs`,
      `--evaluator_*`, `--max_statements_to_execute` as a group.
- [ ] `--output_mode=textproto` / `json` for parse and analyze modes.
      Needs `Serialize(*Proto)` on AST and Resolved node types in
      go-googlesql, or a hand-rolled visitor (significant work).
- [ ] `--descriptor_pool=generated`.
      Needs wasm-boundary descriptor-pool plumbing in go-googlesql so
      the host's `protoreflect.GlobalTypes` (or an equivalent
      registry) can be made available to the analyzer.
- [ ] `--import_path` (IMPORT MODULE).
      Needs `ModuleFactory` exposed by go-googlesql.
- [ ] `--web` and `--port`.
      Naturally follows execute mode; ship after the evaluator is
      available.
- [ ] `DEFINE MACRO` / macro expansion.
      Needs `MacroCatalog` register/lookup methods plus
      `ParserOptions.SetMacroCatalog`. The `MacroCatalog` type is
      already a placeholder in go-googlesql but has no mutator API.
- [ ] `--catalog=tpch_graph`.
      Pure Go-side work: hand-port the property-graph schema
      atop the existing TPCH tables.

## Catalog completeness (pure Go-side work)

- [ ] Extend `catalog/sample.go` with the proto / enum / struct
      tables from upstream `sample_catalog_impl.cc`
      (`TestTable`, `EnumTable`, `ZZZ_AmbiguousHasTestTable`, etc.)
      once Go-side proto descriptor handling is designed.
- [ ] Surface primary keys and per-column nullability through
      `catalog.TableSchema` so DESCRIBE matches upstream output more
      closely. Today the schemas only carry name + type.

## Tooling

- [x] Wire `mise run fmt` through `golangci-lint fmt` (formatters from
      `.golangci.yml`, currently **gofumpt** with `module-path` /
      `extra-rules`).
- [x] Pin `golangci-lint` to a specific minor version in `mise.toml`
      (currently `"2"`); avoid surprises from major upgrades.
- [x] Consider an `actions/cache@v4` step in CI for the wazero
      compilation cache directory, to amortise the ~3 s wasm
      compile across runs.
