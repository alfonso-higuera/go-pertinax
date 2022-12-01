package internal

import "github.com/alfonso-higuera/go-pertinax/collection"

type BaseCollection[V any] struct {
	collection.Collection[V]
}

func (c BaseCollection[V]) NthOrDefault(idx int, defaultValue V) (V, error) {
	if idx <= 0 || idx >= c.Size() {
		return defaultValue, nil
	}
	return c.Nth(idx)
}
