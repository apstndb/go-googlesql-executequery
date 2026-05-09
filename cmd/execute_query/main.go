// Command execute_query is a Go-native port of upstream
// google/googlesql's execute_query tool, layered on top of
// go-googlesql.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	googlesql "github.com/goccy/go-googlesql"

	executequery "github.com/apstndb/go-googlesql-executequery"
	"github.com/apstndb/go-googlesql-executequery/cache"
)

const usage = "Usage: execute_query [flags] {<sql> | -}\n" +
	"  reads SQL from the positional argument, from --sql_file, or stdin (when '-')\n" +
	"\n" +
	"Examples:\n" +
	"  execute_query --mode=parse \"SELECT 1\"\n" +
	"  execute_query --mode=analyze --catalog=sample \"SELECT * FROM KeyValue\"\n"

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		var ufErr *unsupportedExitError
		if errors.As(err, &ufErr) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		if errors.Is(err, errIO) {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(3)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// errIO marks I/O / config errors so main can produce exit code 3.
var errIO = errors.New("io error")

// unsupportedExitError tags errors that should map to exit code 2.
type unsupportedExitError struct{ err error }

func (u *unsupportedExitError) Error() string { return u.err.Error() }
func (u *unsupportedExitError) Unwrap() error { return u.err }

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("execute_query", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		// Best-effort: stderr is the user's terminal in practice, and
		// errcheck on these would just bloat the function for no
		// behaviour change.
		_, _ = fmt.Fprint(stderr, usage)
		_, _ = fmt.Fprintln(stderr, "\nFlags:")
		fs.PrintDefaults()
	}

	rf := registerFlags(fs)

	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := rf.toConfig()
	if err != nil {
		// Distinguish unsupported flag/mode errors so main can
		// pick exit code 2.
		if isUnsupportedErr(err) {
			return &unsupportedExitError{err}
		}
		return err
	}

	sql, err := readSQL(rf.sqlFile, fs.Args(), stdin)
	if err != nil {
		return fmt.Errorf("%w: %v", errIO, err)
	}
	sql = strings.TrimSpace(sql)
	if sql == "" {
		fs.Usage()
		return fmt.Errorf("%w: no SQL provided", errIO)
	}

	if err := setupRuntime(rf); err != nil {
		return err
	}

	w := executequery.NewTextWriter(stdout)
	if err := executequery.Run(context.Background(), sql, cfg, w); err != nil {
		if isUnsupportedErr(err) {
			return &unsupportedExitError{err}
		}
		return err
	}
	return nil
}

func setupRuntime(rf *registeredFlags) error {
	var opts []cache.Option
	if rf.cacheDir != "" {
		opts = append(opts, cache.WithDir(rf.cacheDir))
	}
	if rf.noCache {
		opts = append(opts, cache.Disable())
	}
	switch strings.ToLower(rf.compilationMode) {
	case "", "compiler":
		opts = append(opts, cache.WithCompilationMode(googlesql.CompilationModeCompiler))
	case "interpreter":
		opts = append(opts, cache.WithCompilationMode(googlesql.CompilationModeInterpreter))
	default:
		return fmt.Errorf("unknown --compilation_mode %q", rf.compilationMode)
	}
	return cache.Setup(opts...)
}

func isUnsupportedErr(err error) bool {
	return errors.Is(err, executequery.ErrUnsupportedMode) ||
		errors.Is(err, executequery.ErrUnsupportedSQLMode) ||
		errors.Is(err, executequery.ErrUnsupportedCatalog) ||
		errors.Is(err, executequery.ErrUnsupportedFlag)
}
