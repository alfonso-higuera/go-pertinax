package list

const (
	shiftIncrement = 5
	maxBranches    = 1 << shiftIncrement
	branchMask     = maxBranches - 1
)

type listNode[V any] struct {
	shift    int
	isStrict bool
	numNodes int
	nodes    []any
	offsets  []int
}

func newListNode[V any](shift int) *listNode[V] {
	list := listNode[V]{
		shift:    shift,
		isStrict: false,
		numNodes: 0,
		nodes:    make([]any, 2),
		offsets:  make([]int, 2),
	}

	return &list
}

func newListNodeFromChunk[V any](chunk []V) *listNode[V] {
	return newListNode[V](shiftIncrement).pushChunkLast(chunk)
}

func newListNodeFromNode[V any](shift int, child *listNode[V]) *listNode[V] {
	return newListNode[V](shift).pushNodeLast(child)
}

func newListNodeFromNodes[V any](shift int, a, b *listNode[V]) *listNode[V] {
	return newListNode[V](shift).pushNodeLast(a).pushNodeLast(b)
}

func (l *listNode[V]) size() int {
	if l.numNodes == 0 {
		return 0
	}

	return l.offsets[l.numNodes-1]
}

func (l *listNode[V]) grow() {
	offsets := make([]int, len(l.offsets)<<1)
	copy(offsets, l.offsets)
	l.offsets = offsets

	nodes := make([]any, len(l.nodes)<<1)
	copy(nodes, l.nodes)
	l.nodes = nodes
}

func (l *listNode[V]) offset(idx int) int {
	if idx == 0 {
		return 0
	}

	return l.offsets[idx-1]
}

func (l *listNode[V]) updateStrict() {
	l.isStrict = l.numNodes <= 1 || l.offset(l.numNodes-1) == (l.numNodes-1)*(1<<l.shift)
}

func (l *listNode[V]) pushChunkLast(chunk []V) *listNode[V] {
	if l.size() == 0 && l.shift > shiftIncrement {
		return l.pushNodeLast(newListNodeFromChunk(chunk))
	}

	stack := make([]*listNode[V], l.shift/shiftIncrement)
	stack[0] = l
	for i := 1; i < len(stack); i++ {
		node := stack[i-1]
		stack[i] = node.nodes[node.numNodes-1].(*listNode[V])
	}

	if stack[len(stack)-1].numNodes == maxBranches {
		if l.numNodes == maxBranches {
			return newListNode[V](l.shift + shiftIncrement).pushNodeLast(l).pushChunkLast(chunk)
		}

		return l.pushNodeLast(newListNodeFromChunk(chunk))
	}

	parent := stack[len(stack)-1]
	if len(parent.nodes) == parent.numNodes {
		parent.grow()
	}
	parent.offsets[parent.numNodes] = parent.size()
	parent.numNodes++

	for i, node := range stack {
		lastIdx := node.numNodes - 1
		if i == len(stack)-1 {
			node.nodes[lastIdx] = chunk
		} else {
			node.nodes[lastIdx] = stack[i+1]
		}
		node.offsets[lastIdx] += len(chunk)
		node.updateStrict()
	}

	return stack[0]
}

func (l *listNode[V]) pushNodeFirst(node *listNode[V]) *listNode[V] {
	if l.size() <= 0 {
		panic(l)
	}

	if node.size() == 0 {
		return l
	}

	switch {
	case l.shift < node.shift:
		return node.pushNodeLast(l)
	case l.shift == node.shift:
		return newListNodeFromNodes(l.shift+shiftIncrement, node, l)
	}

	var stackLen int
	if l.numNodes == 0 {
		stackLen = 1
	} else {
		stackLen = (l.shift - node.shift) / shiftIncrement
	}
	stack := make([]*listNode[V], stackLen)
	stack[0] = l
	for i := 1; i < len(stack); i++ {
		n := stack[i-1]
		stack[i] = n.nodes[0].(*listNode[V])
	}

	if stack[len(stack)-1].numNodes == maxBranches {
		return l.pushNodeFirst(newListNodeFromNode(node.shift+shiftIncrement, node))
	}

	parent := stack[len(stack)-1]
	if len(parent.nodes) == parent.numNodes {
		parent.grow()
	}
	copy(parent.nodes[1:parent.numNodes+1], parent.nodes)
	copy(parent.offsets[1:parent.numNodes+1], parent.offsets)
	parent.numNodes++
	parent.offsets[0] = 0

	nodeSize := node.size()
	for i, n := range stack {
		if i == len(stack)-1 {
			n.nodes[0] = node
		} else {
			n.nodes[0] = stack[i+1]
		}
		for j := 0; j < n.numNodes; j++ {
			n.offsets[j] += nodeSize
		}
		n.updateStrict()
	}

	return stack[0]
}

func (l *listNode[V]) pushNodeLast(node *listNode[V]) *listNode[V] {
	if node.size() == 0 {
		return l
	}

	if l.size() == 0 && (l.shift-node.shift) > shiftIncrement {
		return l.pushNodeLast(newListNodeFromNode(node.shift+shiftIncrement, node))
	}

	switch {
	case l.shift < node.shift:
		return node.pushNodeFirst(l)
	case l.shift == node.shift:
		return newListNodeFromNodes(l.shift+shiftIncrement, l, node)
	}

	var stackLen int
	if l.numNodes == 0 {
		stackLen = 1
	} else {
		stackLen = (l.shift - node.shift) / shiftIncrement
	}
	stack := make([]*listNode[V], stackLen)
	stack[0] = l
	for i := 1; i < len(stack); i++ {
		n := stack[i-1]
		stack[i] = n.nodes[n.numNodes-1].(*listNode[V])
	}

	if stack[len(stack)-1].numNodes == maxBranches {
		return l.pushNodeLast(newListNodeFromNode(node.shift+shiftIncrement, node))
	}

	parent := stack[len(stack)-1]
	if len(parent.nodes) == parent.numNodes {
		parent.grow()
	}
	parent.offsets[parent.numNodes] = parent.size()
	parent.numNodes++

	nodeSize := node.size()
	for i, n := range stack {
		lastIdx := n.numNodes - 1
		if i == len(stack)-1 {
			n.nodes[lastIdx] = node
		} else {
			n.nodes[lastIdx] = stack[i+1]
		}
		n.offsets[lastIdx] += nodeSize
		n.updateStrict()
	}

	return stack[0]
}

func (l *listNode[V]) indexOf(idx int) int {
	var estimate int
	if l.shift > 60 {
		estimate = 0
	} else {
		estimate = int((uint(idx) >> l.shift) & branchMask)
	}
	if l.isStrict {
		return estimate
	}

	for i := estimate; i < len(l.nodes); i++ {
		if idx < l.offsets[i] {
			return i
		}
	}

	return -1
}

func (l *listNode[V]) relaxedNthChunk(idx int) ([]V, int) {
	idx = idx & ((1 << (l.shift + shiftIncrement)) - 1)
	node := l
	for node.shift > shiftIncrement {
		nodeIdx := node.indexOf(idx)
		idx -= node.offset(nodeIdx)
		node = node.nodes[nodeIdx].(*listNode[V])
	}

	nodeIdx := node.indexOf(idx)
	return node.nodes[nodeIdx].([]V), node.offset(nodeIdx)
}

func (l *listNode[V]) relaxedNth(idx int) V {
	chunk, offset := l.relaxedNthChunk(idx)
	return chunk[idx-offset]
}

func (l *listNode[V]) nthChunk(idx int) []V {
	if !l.isStrict {
		chunk, _ := l.relaxedNthChunk(idx)
		return chunk
	}

	node := l
	for node.shift > shiftIncrement {
		nodeIdx := int((uint(idx) >> node.shift) & branchMask)
		node = node.nodes[nodeIdx].(*listNode[V])
		if !node.isStrict {
			chunk, _ := node.relaxedNthChunk(idx)
			return chunk
		}
	}

	nodeIdx := int((uint(idx) >> shiftIncrement) & branchMask)
	return node.nodes[nodeIdx].([]V)
}

func (l *listNode[V]) nth(idx int) V {
	if !l.isStrict {
		return l.relaxedNth(idx)
	}

	return l.nthChunk(idx)[idx&branchMask]
}
