package fmtutil

import (
	"fmt"
	"regexp"
)

func PadFormat(f fmt.State, N int) {
	if width, ok := f.Width(); ok {
		if padding := width - N; padding > 0 {
			fmt.Fprintf(f, "%*s", padding, "")
		}
	}
}

var indentRE = regexp.MustCompile("(?m)^")

func IndentBlock(s string) string {
	return indentRE.ReplaceAllLiteralString(s, "    ")
}
