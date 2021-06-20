package fmtutil

import (
	"fmt"
	"strings"
)

func PadFormat(f fmt.State, N int) {
	if width, ok := f.Width(); ok {
		if padding := width - N; padding > 0 {
			fmt.Fprintf(f, "%*s", padding, "")
		}
	}
}

func IndentBlock(s string) string {
	return strings.ReplaceAll(s, "\n", "\n    ")
}
