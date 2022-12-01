package collection

type Collection[V any] interface {
	Iterable[V]

	Size() int
	Nth(idx int) (V, error)
}
