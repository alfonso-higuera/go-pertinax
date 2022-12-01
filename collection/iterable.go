package collection

type Iterable[V any] interface {
	Iterator() Iterator[V]
}
