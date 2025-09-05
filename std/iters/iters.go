// Package iters provides iterators for common use-cases in templates.
//
// The iterators are always deterministic, meaning they always produce the same
// order for the same input data.
// In other words, if the same template with the same input data is rendered
// twice, the output will be exactly the same both times.
// Some iterators require specific properties of the input data to guarantee
// determinism.
package iters

import (
	"cmp"
	"iter"
	"slices"
)

// ============================================================================
// Ordered
// ======================================================================================

// OrderedByKey returns a deterministic iterator over the map's keys and values,
// ordered by its keys.
func OrderedByKey[M ~map[K]V, K cmp.Ordered, V any](m M) iter.Seq2[K, V] {
	keys := keys(m)
	slices.Sort(keys)
	return mapIter(m, keys)
}

// OrderedByValue returns a deterministic iterator over the map's keys and
// values, order by its values.
//
// The returned iterator is only deterministic if all values in the map are
// unique.
// Otherwise, the order of keys with the same value is undefined.
func OrderedByValue[M ~map[K]V, K comparable, V cmp.Ordered](m M) iter.Seq2[K, V] {
	keys := keys(m)
	slices.SortFunc(keys, func(ka, kb K) int {
		return cmp.Compare(m[ka], m[kb])
	})
	return mapIter(m, keys)
}

// MapOrderedBy is the same as [OrderedByValue], but orders the mapping
// of the values as returned by the given mapper.
func MapOrderedBy[M ~map[K]V, K comparable, V any, T cmp.Ordered](m M, mapper func(V) T) iter.Seq2[K, V] {
	keys := keys(m)
	slices.SortFunc(keys, func(ka, kb K) int {
		return cmp.Compare(mapper(m[ka]), mapper(m[kb]))
	})
	return mapIter(m, keys)
}

func keys[M ~map[K]V, K comparable, V any](m M) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func mapIter[M ~map[K]V, K comparable, V any](m M, keys []K) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for _, k := range keys {
			if !yield(k, m[k]) {
				return
			}
		}
	}
}

// Ordered returns a deterministic iterator over the slice's ordered elements.
// The first value of the iterator is the new index of the element, i.e. the
// index the element would have in a sorted slice.
//
// As Ordered does not modify the input slice, it needs to allocate n ints to
// store the sorted indices.
// Consider sorting the slice in-place ([slices.Sort]) and then iterating over
// them regularly for better performance.
func Ordered[S ~[]E, E cmp.Ordered](s S) iter.Seq2[int, E] {
	indices := make([]int, len(s))
	for i := range s {
		indices[i] = i
	}
	slices.SortStableFunc(indices, func(a, b int) int {
		return cmp.Compare(s[a], s[b])
	})
	return func(yield func(int, E) bool) {
		for newIndex, oldIndex := range indices {
			if !yield(newIndex, s[oldIndex]) {
				return
			}
		}
	}
}

// OrderedBy is exactly like [Ordered], but orders the mapping returned by the
// given deterministic mapper function.
//
// OrderedBy is stable, meaning that elements that map to the same key retain
// their original order.
//
// As OrderedBy does not modify the input slice, it needs to allocate n ints to
// store the sorted indices.
// Consider sorting the slice in-place ([slices.SortStableFunc] or
// [slices.Sort] if you know the elements to be unique) and then iterating over
// them regularly for better performance.
func OrderedBy[S ~[]E, E any, K cmp.Ordered](s S, mapper func(E) K) iter.Seq2[int, E] {
	indices := make([]int, len(s))
	for i := range s {
		indices[i] = i
	}
	slices.SortStableFunc(indices, func(a, b int) int {
		return cmp.Compare(mapper(s[a]), mapper(s[b]))
	})
	return func(yield func(int, E) bool) {
		for newIndex, oldIndex := range indices {
			if !yield(newIndex, s[oldIndex]) {
				return
			}
		}
	}
}

// ============================================================================
// Unique
// ======================================================================================

// Unique returns a deterministic iterator in original slice order, filtering
// out duplicate elements.
// That is, for duplicate elements, the returned iterator only yields the
// first appearance.
func Unique[S ~[]E, E comparable](items S) iter.Seq2[int, E] {
	return UniqueUsing(items, func(v E) E { return v })
}

// UniqueUsing is exactly like [Unique], but uses the given deterministic
// mapper function and compares the mapped values instead of the elements
// themselves.
func UniqueUsing[S ~[]E, E any, K comparable](items S, mapper func(E) K) iter.Seq2[int, E] {
	return func(yield func(int, E) bool) {
		seen := make(map[K]struct{})
		for i, item := range items {
			k := mapper(item)
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			if !yield(i, item) {
				return
			}
		}
	}
}

// ============================================================================
// GroupBy
// ======================================================================================

// GroupBy takes a slice of items and a deterministic key function that maps
// each item to a key.
// It returns a deterministic iterator over (key, iter) pairs, where key is a
// unique key, and iter is another deterministic iterator over the items with
// that key.
// The keys appear in order of first appearance in the input slice.
// The nested iterator is stable, meaning that items appear in the same order
// as they appeared in the input slice.
//
// If the input slice is already grouped by the return of the key function,
// it is much more efficient to use [GroupAdjacentBy], to save on the
// additional overhead of identifying and storing the groups.
func GroupBy[S ~[]E, E any, K comparable](items S, key func(E) K) iter.Seq2[K, iter.Seq2[int, E]] {
	groups := make(map[K][]int)
	keys := make([]K, 0, min(32, len(items)/2))
	for i, item := range items {
		k := key(item)
		if groups[k] == nil {
			keys = append(keys, k)
		}
		groups[k] = append(groups[k], i)
	}
	return func(yield func(K, iter.Seq2[int, E]) bool) {
		for _, k := range keys {
			ok := yield(k, func(yield func(int, E) bool) {
				for _, i := range groups[k] {
					if !yield(i, items[i]) {
						return
					}
				}
			})
			if !ok {
				return
			}
		}
	}
}

// GroupItemsBy is exactly like [GroupBy], but iterates over (K, []E) pairs.
//
// Where GroupBy uses a slice of indices to store group members, GroupItemsBy
// needs to create slices of the actual items.
// This can become a problem if iterating over structs (not pointers), as each
// item needs to be copied, essentially duplicating the input slice in memory.
// For pointer types or primitives, GroupItemsBy consumes about as much memory
// as GroupBy.
func GroupItemsBy[S ~[]E, E any, K comparable](items S, key func(E) K) iter.Seq2[K, []E] {
	groups := make(map[K][]E)
	keys := make([]K, 0, min(32, len(items)/2))
	for _, item := range items {
		k := key(item)
		if groups[k] == nil {
			keys = append(keys, k)
		}
		groups[k] = append(groups[k], item)
	}
	return func(yield func(K, []E) bool) {
		for _, k := range keys {
			if !yield(k, groups[k]) {
				return
			}
		}
	}
}

// GroupAdjacentBy is exactly like [GroupBy], but takes advantage of the input
// slice being already grouped by the return of the key function.
func GroupAdjacentBy[S ~[]E, E any, K comparable](items S, key func(E) K) iter.Seq2[K, iter.Seq2[int, E]] {
	return func(yield func(K, iter.Seq2[int, E]) bool) {
		var i int
		for i < len(items) {
			k := key(items[i])
			ok := yield(k, func(yield func(int, E) bool) {
				for i < len(items) && key(items[i]) == k {
					if !yield(i, items[i]) {
						return
					}
					i++
				}
			})
			if !ok {
				return
			}

			// The inner iterator might have not been called at all, or might
			// not have consumed all items of the group.
			// Advance i to the next group.
			for i < len(items) && k == key(items[i]) {
				i++
			}
		}
	}
}
