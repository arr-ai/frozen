# Stability

Once a project reaches 1.0, backwards compatibility becomes a binding
commitment. Breaking changes to the public API require forking into a new
product (e.g. `frozen2`). This document catalogues the interaction surface
so that each release can be mechanically audited for breakage.

## Interaction surface catalogue

Snapshot as of v1.10.0.

### Package `frozen`

#### Interfaces

```go
type Iterator[T any] interface {
    Next() bool
    Value() T
}

type Hashable interface {
    Hash(seed uintptr) uintptr
}

type Key[T any] interface {
    value.Equaler[T]  // Equal(T) bool
    Hashable          // Hash(seed uintptr) uintptr
}
```

#### `Set[T any]`

```go
type Set[T any] struct { /* unexported */ }

func NewSet[T any](values ...T) Set[T]
func Iota[T ~int](stop T) Set[T]
func Iota2[T ~int](start, stop T) Set[T]
func Iota3[T ~int](start, stop, step T) Set[T]
func NewSetFromMask64(mask uint64) Set[int]
func Powerset[T any](s Set[T]) Set[Set[T]]
func Intersection[T any](sets ...Set[T]) Set[T]
func Union[T any](sets ...Set[T]) Set[T]
func SetMap[T, U any](s Set[T], f func(elem T) U) Set[U]
func SetAs[U, T any](s Set[T]) Set[U]
func SetGroupBy[T, K any](s Set[T], key func(el T) K) Map[K, Set[T]]

func (s Set[T]) IsEmpty() bool
func (s Set[T]) Count() int
func (s Set[T]) Range() Iterator[T]
func (s Set[T]) Elements() []T
func (s Set[T]) OrderedElements(less tree.Less[T]) []T
func (s Set[T]) Any() T
func (s Set[T]) AnyN(n int) Set[T]
func (s Set[T]) OrderedFirstN(n int, less tree.Less[T]) []T
func (s Set[T]) First(less tree.Less[T]) any
func (s Set[T]) FirstN(n int, less tree.Less[T]) Set[T]
func (s Set[T]) OrderedRange(less tree.Less[T]) Iterator[T]
func (s Set[T]) Has(val T) bool
func (s Set[T]) With(v T) Set[T]
func (s Set[T]) Without(v T) Set[T]
func (s Set[T]) Where(pred func(elem T) bool) Set[T]
func (s Set[T]) Reduce(reduce func(elems ...T) T) (T, bool)
func (s Set[T]) Reduce2(reduce func(a, b T) T) (T, bool)
func (s Set[T]) Intersection(t Set[T]) Set[T]
func (s Set[T]) Union(t Set[T]) Set[T]
func (s Set[T]) Difference(t Set[T]) Set[T]
func (s Set[T]) SymmetricDifference(t Set[T]) Set[T]
func (s Set[T]) IsSubsetOf(t Set[T]) bool
func (s Set[T]) Equal(t Set[T]) bool
func (s Set[T]) Same(a any) bool
func (s Set[T]) Hash(seed uintptr) uintptr
func (s Set[T]) AsSetAny() Set[any]
func (s Set[T]) String() string
func (s Set[T]) Format(f fmt.State, verb rune)
func (s Set[T]) MarshalJSON() ([]byte, error)
```

Note: `tree.Less[T]` is `func(a, b T) bool` (from internal package;
callers pass function literals without importing the type).

#### `SetBuilder[T any]`

```go
type SetBuilder[T any] struct { /* unexported */ }

func NewSetBuilder[T any](capacity int) *SetBuilder[T]

func (b *SetBuilder[T]) Count() int
func (b *SetBuilder[T]) Add(v T)
func (b *SetBuilder[T]) Remove(v T)
func (b *SetBuilder[T]) Has(v T) bool
func (b *SetBuilder[T]) Finish() Set[T]
func (b SetBuilder[T]) String() string
func (b SetBuilder[T]) Format(f fmt.State, verb rune)
```

#### `Map[K any, V any]`

```go
type Map[K any, V any] struct { /* unexported */ }

func NewMap[K any, V any](kvs ...KeyValue[K, V]) Map[K, V]
func NewMapFromKeys[K any, V any](keys Set[K], f func(key K) V) Map[K, V]
func NewMapFromGoMap[K comparable, V any](m map[K]V) Map[K, V]
func MapMap[K, V, U any](m Map[K, V], f func(key K, val V) U) Map[K, U]
func MapToGoMap[K comparable, V any](m Map[K, V]) map[K]V

func (m Map[K, V]) IsEmpty() bool
func (m Map[K, V]) Count() int
func (m Map[K, V]) Any() (key K, value V)
func (m Map[K, V]) Has(key K) bool
func (m Map[K, V]) Get(key K) (_ V, _ bool)
func (m Map[K, V]) MustGet(key K) V
func (m Map[K, V]) GetElse(key K, deflt V) V
func (m Map[K, V]) GetElseFunc(key K, deflt func() V) V
func (m Map[K, V]) With(key K, val V) Map[K, V]
func (m Map[K, V]) Without(key K) Map[K, V]
func (m Map[K, V]) Keys() Set[K]
func (m Map[K, V]) Values() Set[V]
func (m Map[K, V]) Project(keys ...K) Map[K, V]
func (m Map[K, V]) Where(pred func(key K, val V) bool) Map[K, V]
func (m Map[K, V]) Merge(n Map[K, V], resolve func(key K, a, b V) V) Map[K, V]
func (m Map[K, V]) Update(n Map[K, V]) Map[K, V]
func (m Map[K, V]) Range() MapIterator[K, V]
func (m Map[K, V]) Equal(n Map[K, V]) bool
func (m Map[K, V]) Same(a any) bool
func (m Map[K, V]) Hash(seed uintptr) uintptr
func (m Map[K, V]) String() string
func (m Map[K, V]) Format(f fmt.State, verb rune)
func (m Map[K, V]) MarshalJSON() ([]byte, error)
```

#### `MapIterator[K any, V any]`

```go
type MapIterator[K any, V any] struct { /* unexported */ }

func (i *MapIterator[K, V]) Next() bool
func (i *MapIterator[K, V]) Key() K
func (i *MapIterator[K, V]) Value() V
func (i *MapIterator[K, V]) Entry() (key K, value V)
```

#### `MapBuilder[K any, V any]`

```go
type MapBuilder[K any, V any] struct { /* unexported */ }

func NewMapBuilder[K any, V any](capacity int) *MapBuilder[K, V]

func (b *MapBuilder[K, V]) Count() int
func (b *MapBuilder[K, V]) Put(key K, value V)
func (b *MapBuilder[K, V]) Remove(key K)
func (b *MapBuilder[K, V]) Has(key K) bool
func (b *MapBuilder[K, V]) Get(key K) (V, bool)
func (b *MapBuilder[K, V]) Finish() Map[K, V]
```

#### `IntSet[I integer]`

```go
// integer constraint (unexported): ~int | ~int8 | ... | ~uintptr
type IntSet[I integer] struct { /* unexported */ }

func NewIntSet[I integer](is ...I) IntSet[I]

func (s IntSet[I]) IsEmpty() bool
func (s IntSet[I]) Count() int
func (s IntSet[I]) Range() Iterator[I]
func (s IntSet[I]) Elements() []I
func (s IntSet[I]) Any() I
func (s IntSet[I]) Has(val I) bool
func (s IntSet[I]) With(i I) IntSet[I]
func (s IntSet[I]) Without(i I) IntSet[I]
func (s IntSet[I]) Where(pred func(elem I) bool) IntSet[I]
func (s IntSet[I]) Map(f func(elem I) I) IntSet[I]
func (s IntSet[I]) Intersection(t IntSet[I]) IntSet[I]
func (s IntSet[I]) Union(t IntSet[I]) IntSet[I]
func (s IntSet[I]) IsSubsetOf(t IntSet[I]) bool
func (s IntSet[I]) Equal(t IntSet[I]) bool
func (s IntSet[I]) EqualSet(t IntSet[I]) bool  // Deprecated: use Equal
func (s IntSet[I]) Same(t any) bool
func (s IntSet[I]) Hash(seed uintptr) uintptr
func (s IntSet[I]) String() string
func (s IntSet[I]) Format(f fmt.State, verb rune)
```

#### `KeyValue[K, V any]`

```go
type KeyValue[K, V any] struct {
    Key   K
    Value V
}

func KV[K, V any](k K, v V) KeyValue[K, V]

func (kv KeyValue[K, V]) Hash(seed uintptr) uintptr
func (kv KeyValue[K, V]) Equal(kv2 KeyValue[K, V]) bool
func (kv KeyValue[K, V]) Same(a any) bool
func (kv KeyValue[K, V]) String() string
func (kv KeyValue[K, V]) Format(f fmt.State, verb rune)
```

#### `BitIterator`

```go
type BitIterator uintptr

func (b BitIterator) Next() BitIterator
func (b BitIterator) Index() int
func (b BitIterator) Count() int
func (b BitIterator) Has(i int) bool
func (b BitIterator) With(i int) BitIterator
func (b BitIterator) Without(i int) BitIterator
func (b BitIterator) String() string
```

### Package `lazy`

```go
type Predicate func(el any) bool
type Mapper    func(el any) any

type Set interface {
    IsEmpty() bool
    FastIsEmpty() (empty, ok bool)
    Count() int
    FastCount() (count int, ok bool)
    CountUpTo(limit int) int
    FastCountUpTo(limit int) (count int, ok bool)
    Freeze() Set
    Range() SetIterator
    Hash(seed uintptr) uintptr
    Equal(set any) bool
    EqualSet(set Set) bool
    IsSubsetOf(set Set) bool
    Has(el any) bool
    FastHas(el any) (has, ok bool)
    With(v any) Set
    Without(v any) Set
    Where(pred Predicate) Set
    Map(m Mapper) Set
    Union(set Set) Set
    Intersection(set Set) Set
    Difference(set Set) Set
    SymmetricDifference(set Set) Set
    Powerset() Set
}

type SetIterator interface {
    Next() bool
    Value() any
}

type EmptySet struct{}
// EmptySet implements all Set interface methods.

func Frozen(set frozen.Set[any]) Set
```

### Package `rel` (`pkg/rel`)

```go
type Tuple           = frozen.Map[string, any]
type Relation        = frozen.Set[Tuple]
type RelationBuilder = frozen.SetBuilder[Tuple]

func NewTuple(kvs ...frozen.KeyValue[string, any]) Tuple
func New(header []string, tuples ...[]any) Relation
func Project(s Relation, attrs ...string) Relation
func Join(relations ...Relation) Relation
func CartesianProduct(relations ...Relation) Relation
func Nest(s Relation, attrAttrs frozen.Map[string, frozen.Set[string]]) Relation
func Unnest(s Relation, attr string) Relation
```
