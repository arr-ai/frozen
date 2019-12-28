package frozen

import "strings"

type sides int

const (
	leftSideOnly sides = 1 << iota
	rightSideOnly
	oneSideOnly = leftSideOnly | rightSideOnly
)

func useNeither(_, _ interface{}) interface{} {
	return nil
}

func useRight(a, b interface{}) interface{} {
	return b
}

// matchDelta tracks matching elements.
type matchDelta struct {
	input  int
	output int
}

type composer struct {
	name      string
	delta     *matchDelta
	keep      sides
	mutate    bool
	compose   func(a, b interface{}) interface{}
	calcCount func(counter matchDelta) int
	flipped   *composer
}

func newComposer(
	name string,
	keep sides,
	compose func(a, b interface{}) interface{},
	calcCount func(counter matchDelta) int,
) *composer {
	c := &composer{
		name:      name,
		delta:     &matchDelta{},
		keep:      keep,
		compose:   compose,
		calcCount: calcCount,
	}
	return c
}

func newUnionComposer(abCount int) *composer {
	return newComposer("Union", oneSideOnly, useRight,
		func(counter matchDelta) int { return abCount - 2*counter.input + counter.output },
	)
}

func newSymmetricDifferenceComposer(abCount int) *composer {
	return newComposer("SymmetricDifference", oneSideOnly, useNeither,
		func(counter matchDelta) int { return abCount - 2*counter.input },
	)
}

func newDifferenceComposer(aCount int) *composer {
	return newComposer("Difference", leftSideOnly, useNeither,
		func(counter matchDelta) int { return aCount - counter.input },
	)
}

func (c *composer) count() int {
	return c.calcCount(*c.delta)
}

func (c *composer) flip() *composer {
	if c.flipped != nil {
		c.flipped.flipped = c
		return c.flipped
	}
	d := *c
	d.keep = c.keep&1<<1 | c.keep>>1&1
	d.compose = func(b, a interface{}) interface{} { return c.compose(a, b) }
	d.flipped = c
	return &d
}

func (c *composer) String() string {
	var b strings.Builder
	b.WriteByte('[')
	if c.flipped != nil {
		b.WriteByte('~')
	}
	if c.keep&leftSideOnly != 0 {
		b.WriteByte('<')
	}
	if c.keep&rightSideOnly != 0 {
		b.WriteByte('>')
	}
	b.WriteString(c.name)
	if c.mutate {
		b.WriteByte('!')
	}
	b.WriteByte(']')
	return b.String()
}
