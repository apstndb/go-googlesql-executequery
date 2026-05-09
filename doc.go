// Package executequery is a Go-native port of upstream
// google/googlesql's execute_query tool, layered on top of
// go-googlesql (pure-Go GoogleSQL bindings via wazero).
//
// Today it supports the parse, unparse, and analyze tool modes; the
// unanalyze, explain, and execute modes are recognised at the CLI
// surface but return ErrUnsupportedMode because go-googlesql
// does not yet expose SQLBuilder or a reference evaluator. See
// AGENTS.md for the full feature matrix.
package executequery
