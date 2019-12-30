package frozen

import (
	"log"
	"runtime"
	"sync"
	"testing"

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
	const N = 1000
	arr := make([]interface{}, 0, N)
	for i := 0; i < N; i++ {
		arr = append(arr, i)
	}

	dmp := diffmatchpatch.New()

	work := make(chan func())

	for i := runtime.GOMAXPROCS(0); i > 0; i-- {
		go func() {
			for w := range work {
				w()
			}
		}()
	}

	var wg sync.WaitGroup
	for i := N - 1; i >= 0; i-- {
		i := i
		wg.Add(1)
		work <- func() {
			defer wg.Done()
			corpus := arr[i:]
			if !assertSameElements(t, corpus, NewSet(arr[i:]...).Elements()) {
				failedAt := len(corpus)
				// for {
				var b SetBuilder
				before := b.root.String()
				for j, value := range corpus {
					if j == failedAt {
						t.Log("Set a breakpoint here!")
					}
					b.Add(value)
					after := b.root.String()
					diffs := dmp.DiffMain(before, after, true)
					expected := corpus[:j+1]
					actual := b.root.elements()
					if !assertSameElements(t, expected, actual) {
						expectedOnly, actualOnly := compareElements(expected, actual)
						log.Print("expectedOnly = ", expectedOnly)
						log.Print("actualOnly = ", actualOnly)
						log.Printf("++--\n%v", dmp.DiffPrettyText(diffs))
						// failedAt = j
						break
					}
					before = after
				}
				b.Add(arr[N-1])
				NewSet(arr[i:]...)
				// }
			}
		}
	}

	wg.Wait()
	close(work)
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

func TestSetBuilderWithRedundantAddsAndRemoves(t *testing.T) {
	t.Parallel()

	var b SetBuilder
	for i := 0; i < 35; i++ {
		b.Add(i)
	}
	for i := 10; i < 25; i++ {
		b.Remove(i)
	}
	for i := 5; i < 15; i++ {
		b.Add(i)
	}
	for i := 20; i < 30; i++ {
		b.Remove(i)
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
}
