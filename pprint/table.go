package pprint

import (
	"fmt"
	"io"
	"strings"

	"github.com/dcowgill/toysqleval/eval"
)

func Table(w io.Writer, tab *eval.Table) {
	// Determine the maximum width of each column.
	widths := make([]int, len(tab.Columns))
	for i, col := range tab.Columns {
		widths[i] = len(col.Name)
	}
	for _, row := range tab.Data {
		for i, value := range row {
			widths[i] = maxInt(widths[i], len(valueStr(value)))
		}
	}

	// Create padded-format strings for each column.
	colFmts := make([]string, len(tab.Columns))
	for i, n := range widths {
		colFmts[i] = fmt.Sprintf("%%-%ds", n)
	}

	// Print the column headings.
	{
		sep := " "
		for i, col := range tab.Columns {
			fmt.Fprint(w, sep)
			fmt.Fprintf(w, colFmts[i], col.Name)
			sep = " | "
		}
		fmt.Fprintln(w, "")
	}

	// Print the separator line.
	{
		sep := ""
		for _, n := range widths {
			fmt.Fprint(w, sep)
			fmt.Fprint(w, strings.Repeat("-", n+2))
			sep = "+"
		}
		fmt.Fprintln(w, "")
	}

	// Print the rows.
	for _, row := range tab.Data {
		sep := " "
		for i, value := range row {
			fmt.Fprint(w, sep)
			fmt.Fprintf(w, colFmts[i], valueStr(value))
			sep = " | "
		}
		fmt.Fprintln(w, "")
	}
}

func valueStr(v eval.Value) string {
	if v == nil {
		return ""
	}
	return v.String()
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
