# TODO

Living checklist for follow-up work. Keep entries terse; move detail
into `unsupported.go` reason constants or commit messages.

## Release / housekeeping (post-implementation)

- [x] Rename working directory → `go-googlesql-executequery` (already
      done on disk; this entry was stale).
- [x] Initial commit and push.
      Confirm there are no embedded secrets or stray cache directories
      first (`.tmp/` is gitignored; double-check `git status`).
- [x] Create the GitHub repo at `apstndb/go-googlesql-executequery` and
      push `main` plus the initial release tag.
- [x] Cut the first release: `v0.2.YYYYMMDD` (go-googlesql v0.2.x line).
      See `RELEASING.md` for the carry-over rule. (Shipped `v0.2.20260509`.)
- [x] Confirmed macOS CI leg exercises `cache.Default()`:
      `cache.TestDefaultUnderUserCacheDir` sets `HOME` to a temp dir
      and asserts the resolved path lands under it, which on macOS
      goes through `~/Library/Caches/...`. Run as part of `mise run
      ci` on the `macos-latest` matrix entry.

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
- [x] `--catalog=tpch_graph` is supported. `catalog/tpch_graph.go`
      ports `AddJoinColumns`/`AddJoinColumn`/`AddOneJoinColumn` from
      upstream `tpch_catalog.cc:331-468`, using
      `TypeFactory.MakeRowType3` and `NewSimpleColumn(…,
      isPseudoColumn=true, isWritableColumn=false)`. Analyze and
      DESCRIBE both work; pseudo-col traversal (`SELECT c.C_NAME FROM
      Customer c, c.Orders o`) resolves correctly when the user
      passes `--enabled_language_features=ALL_MINUS_DEV,+FEATURE_ROW_TYPE`
      (mirrors upstream's requirement, called out in `--help`).

      Found and fixed a goccy v0.2.1 footgun while wiring this:
      `ParserOptions.SetLanguageOptions` *moves-from* its argument on
      the wasm side, leaving the caller's `*LanguageOptions` handle
      pointing at a default-constructed instance with no features
      enabled. `executequery.go:55-61` now builds a dedicated
      `parserLO` so the analyzer's LO survives unchanged. File this
      upstream against `goccy/go-googlesql` so the move-from semantics
      are documented (or fixed to copy).

      Known residual gap: `OptionalJoinColumnAttributes` is exported
      by goccy v0.2.1 but with no constructor or `SimpleColumn`
      setter, so we cannot attach the upstream
      `Column::JoinColumnAttributes`. Pseudo-col walks resolve via the
      RowType, but anything that depends on JoinColumnAttributes (e.g.
      upstream's join-flattening rewrite, evaluator-side multi-row
      expansion) will diverge from upstream behaviour.

## Catalog completeness (pure Go-side work)

- [~] Proto / enum / struct sample tables — design complete and a
      demo subset shipped. `catalog/sample_proto.go` hand-builds a
      `descriptorpb.FileDescriptorProto` for `zetasql_test`
      (`TestEnum` + `KitchenSinkPB` with `int64_val` / `string_val` /
      `test_enum`), wires it through `goccy.NewDescriptorPool` →
      `BuildFile` → `FindMessageTypeByName` /
      `FindEnumTypeByName` → `TypeFactory.MakeProtoType` /
      `MakeEnumType`, then registers `TestTable` and `EnumTable` with
      proto- and enum-typed columns. DESCRIBE prints
      `PROTO<…>` / `ENUM<…>` labels; analyze of `KitchenSink.int64_val`
      resolves through `GetProtoField`. The remaining upstream tables
      (`ZZZ_AmbiguousHasTestTable`, `Proto3Table`, `MapFieldTable`,
      `CivilTimeTestTable`, `FieldFormatsTable`, etc.) are not yet
      ported — bringing them in needs the corresponding upstream
      `*.proto` definitions, which would expand the
      `third_party/googlesql` sparse checkout. Track as follow-up.
- [x] Primary keys: `TableSchema` now carries `PrimaryKey []string`,
      `buildTable` calls `SimpleTable.SetPrimaryKey`, and
      `Format()` prints `Primary key: (...)`. tpch_schema and the
      keyed sample tables (`KeyValueWithPrimaryKey`, `TwoIntegers`,
      `FourIntegers`) populate it. Per-column nullability is not yet
      surfaced — `NewSimpleColumn` has no nullable bool; would need
      `NewSimpleColumn2(*AnnotatedType, …)` plus an `AnnotatedType`
      builder. Defer until needed.

## Tooling

- [x] Wire `mise run fmt` through `golangci-lint fmt` (formatters from
      `.golangci.yml`, currently **gofumpt** with `module-path` /
      `extra-rules`).
- [x] Pin `golangci-lint` to a specific minor version in `mise.toml`
      (currently `"2"`); avoid surprises from major upgrades.
- [x] Consider an `actions/cache@v4` step in CI for the wazero
      compilation cache directory, to amortise the ~3 s wasm
      compile across runs.
