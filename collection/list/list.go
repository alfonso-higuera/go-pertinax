package list

import (
	"fmt"

	"github.com/alfonso-higuera/go-pertinax/collection/internal"
)

const maxChunkSize = maxBranches

type List[V any] struct {
	prefixLen int
	prefix    []V
	suffixLen int
	suffix    []V
	isLinear  bool
	root      *listNode[V]

	internal.BaseCollection[V]
}

func newList[V any](isLinear bool, root *listNode[V], prefixLen int, prefix []V, suffixLen int, suffix []V) *List[V] {
	list := &List[V]{
		prefixLen: prefixLen,
		prefix:    prefix,
		suffixLen: suffixLen,
		suffix:    suffix,
		isLinear:  isLinear,
		root:      root,
	}
	list.BaseCollection = internal.BaseCollection[V]{Collection: list}

	return list
}

func NewList[V any](vs ...V) *List[V] {
	list := newList(true, newListNode[V](shiftIncrement), 0, nil, 0, nil)

	for _, v := range vs {
		list.AddLast(v)
	}

	return list.Forked()
}

func (l *List[V]) Linear() *List[V] {
	if l.IsLinear() {
		return l
	}

	return newList(true, l.root, l.prefixLen, l.prefix, l.suffixLen, l.suffix).Clone()
}

func (l *List[V]) Forked() *List[V] {
	if l.IsLinear() {
		return newList(false, l.root, l.prefixLen, l.prefix, l.suffixLen, l.suffix).Clone()
	}

	return l
}

func (l *List[V]) growSuffix() {
	increasedSuffixLen := len(l.suffix) << 1
	var newSuffixLen int
	if increasedSuffixLen < maxChunkSize {
		newSuffixLen = increasedSuffixLen
	} else {
		newSuffixLen = maxChunkSize
	}

	newSuffix := make([]V, newSuffixLen)
	copy(newSuffix, l.suffix)

	l.suffix = newSuffix
}

func (l *List[V]) pushLast(value V) *List[V] {
	switch {
	case l.suffix == nil:
		l.suffix = make([]V, 2)
	case l.suffixLen == len(l.suffix):
		l.growSuffix()
	}

	l.suffix[l.suffixLen] = value
	l.suffixLen++

	if l.suffixLen == maxChunkSize {
		l.root = l.root.pushChunkLast(l.suffix)
		l.suffix = nil
		l.suffixLen = 0
	}

	return l
}

func (l *List[V]) IsLinear() bool {
	return l.isLinear
}

func (l *List[V]) AddLast(value V) *List[V] {
	if l.IsLinear() {
		return l.pushLast(value)
	}

	return l.Clone().pushLast(value)
}

func (l *List[V]) Clone() *List[V] {
	var newPrefix []V
	if l.prefix != nil {
		newPrefix = make([]V, len(l.prefix))
		copy(newPrefix, l.prefix)
	}

	var newSuffix []V
	if l.suffix != nil {
		newSuffix = make([]V, len(l.suffix))
		copy(newSuffix, l.suffix)
	}

	return newList(l.IsLinear(), l.root, l.prefixLen, newPrefix, l.suffixLen, newSuffix)
}

func (l *List[V]) Size() int {
	return l.root.size() + l.prefixLen + l.suffixLen
}

func (l *List[V]) Nth(idx int) (V, error) {
	rootSize := l.root.size()
	if idx < 0 || idx >= (rootSize+l.prefixLen+l.suffixLen) {
		return *new(V), fmt.Errorf("%d must be within [0,%d)", idx, l.Size())
	}

	switch {
	case idx < l.prefixLen:
		return l.prefix[len(l.prefix)+idx-l.prefixLen], nil
	case idx-l.prefixLen < rootSize:
		return l.root.nth(idx - l.prefixLen), nil
	default:
		return l.suffix[idx-(rootSize+l.prefixLen)], nil
	}
}
