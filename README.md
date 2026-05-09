# go-googlesql-executequery

A Go-native port of upstream
[`google/googlesql`](https://github.com/google/googlesql)'s
[`execute_query`](https://github.com/google/googlesql/blob/master/execute_query.md)
tool, built on top of
[`goccy/go-googlesql`](https://github.com/goccy/go-googlesql) (pure-Go
GoogleSQL bindings via wazero — no cgo).

This repository ships:

- `execute_query` — a CLI binary that mirrors the upstream flag surface.
- `executequery` — an exported Go package usable as a library.
- `executequery/cache` — a small helper that resolves a safe,
  per-user, per-`go-googlesql`-version wazero compilation-cache directory.

## Status

`goccy/go-googlesql` only exposes parser and analyzer functionality;
its compiled `googlesql.wasm` does not yet ship a reference evaluator
or a Resolved-AST → SQL builder. As a result this port supports three
of the upstream tool's six modes today:

| Mode        | Status        |
|-------------|---------------|
| `parse`     | supported     |
| `unparse`   | supported     |
| `analyze`   | supported     |
| `unanalyze` | not supported (no `SQLBuilder` exposed by `goccy/go-googlesql`) |
| `explain`   | not supported (no reference evaluator exposed by `goccy/go-googlesql`) |
| `execute`   | not supported (no reference evaluator exposed by `goccy/go-googlesql`) |

Every upstream `execute_query` flag is recognised by the CLI; flags
that depend on the unsupported modes return a structured error rather
than `flag provided but not defined`. The Go source-code declaration
of each unsupported flag carries a comment explaining what it does in
upstream, why it cannot be honoured today, and what change unblocks it.

## Versioning

Releases use `v<go-googlesql-Major.Minor>.<YYYYMMDD>`, with date
carry-over on collision. See `RELEASING.md`.

## Quick start

```sh
mise install                       # installs Go and golangci-lint
git submodule update --init        # populates third_party/googlesql
mise run build                     # produces bin/execute_query

bin/execute_query --mode=parse 'SELECT 1+1'
bin/execute_query --mode=analyze --catalog=tpch 'SELECT count(*) FROM Orders'
bin/execute_query --catalog=tpch 'DESCRIBE Orders'
```

`mise run ci` runs lint, tests, and build.

## Library

```go
import (
    "context"
    "os"

    "github.com/apstndb/go-googlesql-executequery"
    "github.com/apstndb/go-googlesql-executequery/cache"
)

func main() {
    if err := cache.Setup(); err != nil {
        panic(err)
    }
    cfg := executequery.Config{Modes: []executequery.Mode{executequery.ModeAnalyze}}
    if err := executequery.Run(context.Background(), "SELECT 1+1",
        cfg, executequery.NewTextWriter(os.Stdout)); err != nil {
        panic(err)
    }
}
```

## License

Apache-2.0. See `LICENSE` and `NOTICE`. This is a derivative work of
`google/googlesql` (Apache-2.0); upstream files referenced under
`third_party/googlesql/` are imported via a git submodule pinned to a
specific commit.
