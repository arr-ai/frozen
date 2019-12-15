package frozen

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

type composer struct {
	middleInCell  int
	middleOutCell int
	middleIn      *int
	middleOut     *int
	keep          sides
	mutate        bool
	compose       func(a, b interface{}) interface{}
	calcCount     func(middleIn, middleOut int) int
	flipped       *composer
}

func newComposer(
	keep sides,
	compose func(a, b interface{}) interface{},
	calcCount func(middleIn, middleOut int) int,
) *composer {
	c := &composer{
		keep:      keep,
		compose:   compose,
		calcCount: calcCount,
	}
	c.middleIn = &c.middleInCell
	c.middleOut = &c.middleOutCell
	return c
}

func newIntersectionComposer() *composer {
	return newComposer(0, useRight,
		func(middleIn, middleOut int) int { return middleOut },
	)
}

func newUnionComposer(abCount int) *composer {
	return newComposer(oneSideOnly, useRight,
		func(middleIn, middleOut int) int { return abCount - 2*middleIn + middleOut },
	)
}

func newSymmetricDifferenceComposer(abCount int) *composer {
	return newComposer(oneSideOnly, useNeither,
		func(middleIn, middleOut int) int { return abCount - 2*middleIn + middleOut },
	)
}

func newMinusComposer(aCount int) *composer {
	return newComposer(leftSideOnly, useNeither,
		func(middleIn, middleOut int) int { return aCount - middleIn + middleOut },
	)
}

func (c *composer) count() int {
	return c.calcCount(*c.middleIn, *c.middleOut)
}

func (c *composer) flip() *composer {
	if c.flipped != nil {
		c.flipped.flipped = c
		return c.flipped
	}
	return &composer{
		middleIn:  c.middleIn,
		middleOut: c.middleOut,
		keep:      c.keep&1<<1 | c.keep>>1&1,
		compose:   func(b, a interface{}) interface{} { return c.compose(a, b) },
		calcCount: c.calcCount,
		flipped:   c,
	}
}
