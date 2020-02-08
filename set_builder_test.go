package frozen

import (
	"log"
	"testing"

	"github.com/arr-ai/frozen/types"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stretchr/testify/assert"
)

func TestSetBuilderEmpty(t *testing.T) {
	t.Parallel()

	var b SetBuilder
	assertSetEqual(t, Set{}, b.Finish())
}

func TestSetBuilder(t *testing.T) {
	t.Parallel()

	var b SetBuilder
	for i := 0; i < 10; i++ {
		b.Add(i)
	}
	m := b.Finish()
	assert.Equal(t, 10, m.Count())
	for i := 0; i < 10; i++ {
		assert.True(t, m.Has(i))
	}
}

func TestSetBuilderIncremental(t *testing.T) {
	replayable(true, func(r replayer) {
		const N = 1000
		arr := make([]interface{}, 0, N)
		for i := 0; i < N; i++ {
			arr = append(arr, i)
		}

		dmp := diffmatchpatch.New()

		for i := N - 1; i >= 0; i-- {
			i := i
			corpus := arr[i:]
			if !assertSameElements(t, corpus, NewSet(arr[i:]...).Elements()) {
				var b SetBuilder
				for j, value := range corpus {
					var before string
					if r.mark(i, j).isTarget {
						before = b.root.String()
						log.Printf("before = %v", before)
						func() {
							// defer logrus.SetLevel(logrus.GetLevel())
							// logrus.SetLevel(logrus.TraceLevel)
							b.Add(value)
						}()
					} else {
						b.Add(value)
					}
					expected := corpus[:j+1]
					actual := b.root.Elements(0)
					if !assertSameElements(t, expected, actual) {
						after := b.root.String()
						log.Printf("after = %v", after)
						diffs := dmp.DiffMain(before, after, false)
						log.Printf("++--\n%v", dmp.DiffPrettyText(diffs))
						expectedOnly, actualOnly := compareElements(expected, actual)
						log.Print("expectedOnly = ", expectedOnly)
						log.Print("actualOnly = ", actualOnly)
						r.replay()
					}
				}
				b.Add(arr[N-1])
				NewSet(arr[i:]...)
			}
		}
	})
}

func TestSetBuilderRemove(t *testing.T) {
	t.Parallel()

	var b SetBuilder
	for i := 0; i < 15; i++ {
		b.Add(i)
	}
	for i := 5; i < 10; i++ {
		b.Remove(i)
	}
	m := b.Finish()

	assert.Equal(t, 10, m.Count())
	for i := 0; i < 15; i++ {
		switch {
		case i < 5:
			assertSetHas(t, m, i)
		case i < 10:
			assertSetNotHas(t, m, i)
		default:
			assertSetHas(t, m, i)
		}
	}
}

func TestSetBuilderWithRedundantAddsAndRemoves(t *testing.T) { //nolint:funlen
	t.Parallel()

	replayable(false, func(r replayer) {
		var b SetBuilder

		s := uint64(0)

		requireMatch := func(format string, args ...interface{}) {
			for j := 0; j < 35; j++ {
				if !assert.Equalf(t, s&(uint64(1)<<j) != 0, b.Has(j), format+" j=%v", append(args, j)...) {
					log.Print(s&(uint64(1)<<j) != 0, b.Has(j), types.BitIterator(s), b.root)
					b.Has(j)
					r.replay()
					t.FailNow()
				}
			}
		}

		add := func(i int) {
			b.Add(i)
			s |= uint64(1) << i
		}

		remove := func(i int) {
			b.Remove(i)
			s &^= uint64(1) << i
		}

		requireMatch("")
		for i := 0; i < 35; i++ {
			add(i)
			requireMatch("i=%v", i)
		}
		for i := 10; i < 25; i++ {
			remove(i)
			requireMatch("i=%v", i)
		}

		for i := 5; i < 15; i++ {
			if r.mark(i).isTarget {
				log.Print(types.BitIterator(s), b.root)
			}
			add(i)
			requireMatch("i=%v", i)
		}
		for i := 20; i < 30; i++ {
			remove(i)
			requireMatch("i=%v", i)
		}
		m := b.Finish()

		for i := 0; i < 35; i++ {
			switch {
			case i < 15:
				assertSetHas(t, m, i)
			case i < 30:
				assertSetNotHas(t, m, i)
			default:
				assertSetHas(t, m, i)
			}
		}
	})
}
