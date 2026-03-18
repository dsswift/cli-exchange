package output

import (
	"fmt"
	"strings"
)

func BuildTable(headers []string, rows [][]string) string {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	sep := "  "
	var b strings.Builder

	// Header
	for i, h := range headers {
		if i > 0 {
			b.WriteString(sep)
		}
		fmt.Fprintf(&b, "%-*s", widths[i], h)
	}
	b.WriteString("\n")

	// Divider
	for i, w := range widths {
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteString(strings.Repeat("-", w))
	}
	b.WriteString("\n")

	// Rows
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				b.WriteString(sep)
			}
			if i < len(widths) {
				fmt.Fprintf(&b, "%-*s", widths[i], cell)
			} else {
				b.WriteString(cell)
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
