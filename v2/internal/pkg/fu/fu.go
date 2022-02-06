package fu

import (
	"fmt"
	"io"
	"strings"
)

func PadFormat(f fmt.State, N int) {
	if width, ok := f.Width(); ok {
		if padding := width - N; padding > 0 {
			fmt.Fprintf(f, "%*s", padding, "")
		}
	}
}

func Sep(w io.Writer, i int, sep string) bool {
	if i > 0 {
		Fprint(w, sep)
		return true
	}
	return false
}

func Comma(w io.Writer, i int) bool {
	return Sep(w, i, ", ")
}

func IndentBlock(s string) string {
	return strings.ReplaceAll(s, "\n", "\n    ")
}

func Format(i any, f fmt.State, verb rune) {
	switch i := i.(type) {
	case fmt.Formatter:
		i.Format(f, verb)
	default:
		Fprint(f, i)
	}
}

func Fprint(w io.Writer, a ...any) {
	if _, err := fmt.Fprint(w, a...); err != nil {
		panic(err)
	}
}

func Fprintf(w io.Writer, format string, a ...any) {
	if _, err := fmt.Fprintf(w, format, a...); err != nil {
		panic(err)
	}
}

func String(i any) string {
	return fmt.Sprintf("%s", i)
}

func WriteString(w io.Writer, s string) {
	if _, err := io.WriteString(w, s); err != nil {
		panic(err)
	}
}
