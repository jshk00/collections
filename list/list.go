package list

type Node[V any] struct {
	next, prev *Node[V]
	list       *List[V]
	Value      V
}

func (n *Node[V]) Next() *Node[V] {
	if p := n.next; n.list != nil && p != &n.list.root {
		return p
	}
	return nil
}

func (n *Node[V]) Prev() *Node[V] {
	if p := n.prev; n.list != nil && p != &n.list.root {
		return p
	}
	return nil
}

type List[V any] struct {
	root Node[V]
	len  int
}

func New[V any]() *List[V] {
	l := &List[V]{}
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	return l
}

func (l *List[V]) Front() *Node[V] {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

func (l *List[V]) Back() *Node[V] {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

func (l *List[V]) insert(n, at *Node[V]) *Node[V] {
	n.prev = at
	n.next = at.next
	n.prev.next = n
	n.next.prev = n
	n.list = l
	l.len++
	return n
}

func (l *List[V]) Insert(v V, n *Node[V]) *Node[V] {
	return l.insert(&Node[V]{Value: v}, n)
}

func (l *List[V]) PushFront(v V) *Node[V] {
	return l.Insert(v, &l.root)
}

func (l *List[V]) PushBack(v V) *Node[V] {
	return l.Insert(v, l.root.prev)
}

func (l *List[V]) InsertBefore(v V, n *Node[V]) *Node[V] {
	if l != n.list || n == nil {
		return nil
	}
	return l.Insert(v, n.prev)
}

func (l *List[V]) InsertAfter(v V, n *Node[V]) *Node[V] {
	if l != n.list || n == nil {
		return nil
	}
	return l.Insert(v, n)
}

func (l *List[V]) MoveToFront(n *Node[V]) {
	if n.list != l || l.root.next == n {
		return
	}
	l.move(n, &l.root)
}

func (l *List[V]) MoveToBack(n *Node[V]) {
	if n.list != l || l.root.next == n {
		return
	}
	l.move(n, l.root.prev)
}

func (l *List[V]) MoveBefore(n, mark *Node[V]) {
	if n.list != l || n == mark || mark.list != l {
		return
	}
	l.move(n, mark.prev)
}

func (l *List[V]) MoveAfter(n, mark *Node[V]) {
	if n.list != l || n == mark || mark.list != l {
		return
	}
	l.move(n, mark)
}

func (l *List[V]) Remove(n *Node[V]) any {
	if n.list == l {
		l.remove(n)
	}
	return n.Value
}

func (l *List[V]) Len() int {
	return l.len
}

func (l *List[V]) remove(n *Node[V]) {
	n.prev.next = n.next
	n.next.prev = n.prev
	n.prev = nil
	n.next = nil
	n.list = nil
	l.len--
}

func (l *List[V]) move(n, at *Node[V]) {
	if n == at {
		return
	}
	n.prev.next = n.next
	n.next.prev = n.prev
	n.prev = at
	n.next = at.next
	n.prev.next = n
	n.next.prev = n
}
