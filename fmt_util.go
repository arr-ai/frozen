package frozen

import (
	"fmt"
	"regexp"
)

func padFormat(f fmt.State, N int) {
	if width, ok := f.Width(); ok {
		if padding := width - N; padding > 0 {
			fmt.Fprintf(f, "%*s", padding, "")
		}
	}
}

var indentRE = regexp.MustCompile("(?m)^")

func indentBlock(s string) string {
	return indentRE.ReplaceAllLiteralString(s, "    ")
}
