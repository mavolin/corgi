package list

import "container/list"

type List[T any] struct {
	l *list.List
}

func List1[T any](v T) List[T] {
	var l List[T]
	l.PushFront(v)
	return l
}

type Element[T any] struct {
	e *list.Element
}

func (e *Element[T]) Prev() *Element[T] {
	if prev := e.e.Prev(); prev != nil {
		return &Element[T]{prev}
	}
	return nil
}

func (e *Element[T]) Next() *Element[T] {
	if next := e.e.Next(); next != nil {
		return &Element[T]{next}
	}
	return nil
}

func (e *Element[T]) V() T { return e.e.Value.(T) }

func (l *List[T]) Len() int { return l.l.Len() }

func (l *List[T]) Front() *Element[T] {
	if front := l.l.Front(); front != nil {
		return &Element[T]{front}
	}
	return nil
}

func (l *List[T]) Back() *Element[T] {
	if back := l.l.Back(); back != nil {
		return &Element[T]{back}
	}
	return nil
}

func (l *List[T]) Remove(e *Element[T]) T    { return l.l.Remove(e.e).(T) }
func (l *List[T]) PushFront(v T) *Element[T] { return &Element[T]{l.l.PushFront(v)} }
func (l *List[T]) PushBack(v T) *Element[T]  { return &Element[T]{l.l.PushBack(v)} }

func (l *List[T]) InsertBefore(v T, mark *Element[T]) *Element[T] {
	return &Element[T]{l.l.InsertBefore(v, mark.e)}
}

func (l *List[T]) InsertAfter(v T, mark *Element[T]) *Element[T] {
	return &Element[T]{l.l.InsertAfter(v, mark.e)}
}

func (l *List[T]) MoveToFront(e *Element[T])      { l.l.MoveToFront(e.e) }
func (l *List[T]) MoveToBack(e *Element[T])       { l.l.MoveToBack(e.e) }
func (l *List[T]) MoveBefore(e, mark *Element[T]) { l.l.MoveBefore(e.e, mark.e) }
func (l *List[T]) MoveAfter(e, mark *Element[T])  { l.l.MoveAfter(e.e, mark.e) }
func (l *List[T]) PushBackList(other *List[T])    { l.l.PushBackList(other.l) }
func (l *List[T]) PushFrontList(other *List[T])   { l.l.PushFrontList(other.l) }

func (l *List[T]) ToSlice() []T {
	s := make([]T, 0, l.Len())
	for e := l.Front(); e != nil; e = e.Next() {
		s = append(s, e.V())
	}
	return s
}
