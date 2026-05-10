package executequery

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// boxUnicodeSingleColumn renders inner text as one boxed column titled columnTitle,
// matching the default --output_mode=box layout from upstream execute_query
// (googlesql/tools/execute_query/output_query_result.cc: Unicode box glyphs).
func boxUnicodeSingleColumn(columnTitle, inner string) string {
	inner = strings.TrimRight(inner, "\r\n")
	lines := strings.Split(inner, "\n")
	all := append([]string{columnTitle}, lines...)
	mw := 0
	for _, s := range all {
		if n := utf8.RuneCountInString(s); n > mw {
			mw = n
		}
	}
	hBar := strings.Repeat("─", mw+2)
	var b strings.Builder
	b.WriteString("┌")
	b.WriteString(hBar)
	b.WriteString("┐\n")
	writeRow := func(line string) {
		b.WriteString("│ ")
		fmt.Fprintf(&b, "%-*s", mw, line)
		b.WriteString(" │\n")
	}
	writeRow(columnTitle)
	if len(lines) > 0 {
		b.WriteString("├")
		b.WriteString(hBar)
		b.WriteString("┤\n")
		for _, line := range lines {
			writeRow(line)
		}
	}
	b.WriteString("└")
	b.WriteString(hBar)
	b.WriteString("┘")
	return b.String()
}
