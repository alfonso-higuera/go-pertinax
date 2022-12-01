package collection

type Iterator[V any] interface {
	HasNext() bool
	Next() V
}
