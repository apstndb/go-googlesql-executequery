package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// readSQL resolves the SQL input given the parsed flags. Order of
// precedence:
//
//  1. --sql_file=path
//  2. positional args, joined with spaces (mirrors upstream's
//     `execute_query "SELECT 1"` behaviour)
//  3. when the lone positional is `-`, or no positional and stdin
//     is not a terminal, read from stdin
//  4. when the lone positional starts with `@`, read from the
//     remainder as a file path
func readSQL(sqlFile string, positional []string, stdin io.Reader) (string, error) {
	if sqlFile != "" {
		b, err := os.ReadFile(sqlFile)
		if err != nil {
			return "", fmt.Errorf("read --sql_file: %w", err)
		}
		return string(b), nil
	}
	switch len(positional) {
	case 0:
		b, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("read stdin: %w", err)
		}
		return string(b), nil
	case 1:
		arg := positional[0]
		switch {
		case arg == "-":
			b, err := io.ReadAll(stdin)
			if err != nil {
				return "", fmt.Errorf("read stdin: %w", err)
			}
			return string(b), nil
		case strings.HasPrefix(arg, "@"):
			b, err := os.ReadFile(arg[1:])
			if err != nil {
				return "", fmt.Errorf("read %q: %w", arg, err)
			}
			return string(b), nil
		}
		return arg, nil
	}
	return strings.Join(positional, " "), nil
}
