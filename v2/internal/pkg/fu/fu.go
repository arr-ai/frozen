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

func IndentBlock(s string) string {
	return strings.ReplaceAll(s, "\n", "\n    ")
}

func Format(i interface{}, f fmt.State, verb rune) {
	switch i := i.(type) {
	case fmt.Formatter:
		i.Format(f, verb)
	default:
		Fprint(f, i)
	}
}

func Fprint(w io.Writer, a ...interface{}) {
	if _, err := fmt.Fprint(w, a...); err != nil {
		panic(err)
	}
}

func Fprintf(w io.Writer, format string, a ...interface{}) {
	if _, err := fmt.Fprintf(w, format, a...); err != nil {
		panic(err)
	}
}

func String(i interface{}) string {
	return fmt.Sprintf("%s", i)
}

func WriteString(w io.Writer, s string) {
	if _, err := io.WriteString(w, s); err != nil {
		panic(err)
	}
}
