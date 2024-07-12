// Package flatset provides the sorted associative containers 'FlatSet' and 'FlatMultiSet' that store data in
// continuous memory instead of a binary tree (similar to C++ std::flat_set and std::flat_multiset).
//
// Both the FlatSet and FlatMultiSet implement stable ordering, but the previous indices for values are invalidated
// following an method that modifies the container.
//
// Compared to a binary tree, flatsets are considerably faster to read as they avoid the cache misses from pointer
// referencing. On modern CPUs it is also surprisingly fast to insert into a flatset, especially if the collection is
// small, or if you frequently update the end of the flatset. For larger collections you can optimize write operations
// by using a flatset of pointers to structures instead of values, although you must be careful not to modify the data
// inside the flatset as it might affect how the data is sorted.
//
// This package requires 'range over functions' so for golang < 1.23 you must build your application with
// 'GOEXPERIMENT=rangefunc'. https://go.dev/wiki/RangefuncExperiment
//
package flatset


import (
    "iter"
    "reflect"
    "sort"
)

// This is the interface for the comparison function that is passed to the FlatSet and FlatMultiSet which defines how
// the data will be sorted. For example, to sort the data in ascending order the comparison function would implement
// less than.
//
type Compare[V any] func(a, b V) bool


// This is base structure that contains the data for both the FlatSet and FlatMultiSet implementations.
//
type base[V any] struct {
    cmp Compare[V]  // comparison function
    data [] V       // data stored in a array of continuous memory
}


// Shared private method to efficiently insert into an array.
//
func (self *base[V]) insert(ub int, value V) {
    if ub == 0 {
        self.data = append([]V{value}, self.data...)
	} else if ub == len(self.data) {
	    self.data = append(self.data, []V{value}...)
	} else {
     	self.data = append(self.data[:ub], self.data[ub - 1:]...)
    	self.data[ub] = value
	}
}


// Shared private method to search for an value in O(log n) operations using a comparison function.
//
func (self *base[V]) bounds(value V, low int, high int, cmp Compare[V]) int {
	for low <= high {
		mid := (low + high) / 2
		if cmp(self.data[mid], value) {
    		low = mid + 1
		} else {
    		high = mid - 1
		}
	}
	return low
}


// Shared private method that searches for several values with in an array using an iterator. The location of previous
// values are used to optimize the search for the next value. As consecutive values are likely to be in a similar range,
// this algorithm will typically out perform the O(log n) complexity required to search for the values individually.
//
//
func (self *base[V]) traverse(values iter.Seq[V], cmp Compare[V]) iter.Seq2[int, V] {
    low, high := 0, len(self.data) - 1
    idx := (low + high) / 2

    return func(yield func(int, V) bool) {
        for value := range values {
            size := len(self.data)
            if size > 0 {
                if idx == size || self.cmp(value, self.data[idx]) {
                    if self.cmp(value, self.data[low]) {
                        high = low - 1
                        low = 0
                    } else {
                        high = idx - 1
                    }
                } else if high >= 0 {
                    if self.cmp(value, self.data[high]) {
                        low = idx
                        high--
                    } else {
                        low = high
                        high = size - 1
                    }
                } else {
                    low = idx
                    high = size - 1
                }

                idx = self.bounds(value, low, high, cmp)
                if !yield(idx, value) {
                    break
                }
            }
        }
	}
}


// Shared private method to append another flatset to this one that is sorted using the same comparison function.
//
func (self *base[V]) mergeSorted(other *base[V]) {
    lhsIdx, rhsIdx, mergedIdx := 0, 0, 0
    lhsSz, rhsSz := len(self.data), len(other.data)
    mergedSz := lhsSz + rhsSz
    data := make([]V, mergedSz)

    for lhsIdx < lhsSz && rhsIdx < rhsSz {
        if self.cmp(self.data[lhsIdx], other.data[rhsIdx]) {
            data[mergedIdx] = self.data[lhsIdx]
            lhsIdx++
        } else {
            data[mergedIdx] = other.data[rhsIdx]
            rhsIdx++
        }
        mergedIdx++
    }

    if lhsIdx < lhsSz {
        copy(data[mergedIdx:mergedSz], self.data[lhsIdx:lhsSz])
    } else {
        copy(data[mergedIdx:mergedSz], other.data[rhsIdx:rhsSz])
    }
    self.data = data
}


// Returns a copy of the value at the given index.
//
func (self *base[V]) At(index int) V {
    return self.data[index]
}


// Returns the number of values stored in this container.
//
func (self *base[V]) Size() int {
    return len(self.data)
}


// Returns an iterator that returns a copy of each value in order.
//
func (self *base[V]) Values() iter.Seq[V] {
    return func(yield func(V) bool) {
        for i := 0; i < len(self.data); i++ {
            if !yield(self.data[i]) {
                break
            }
        }
    }
}


// Returns an iterator that iterates in reverse order returning a copy of each value.
//
func (self *base[V]) Reversed() iter.Seq[V] {
    return func(yield func(V) bool) {
        for i := len(self.data) - 1; i >= 0; i-- {
            if !yield(self.data[i]) {
                break
            }
        }
    }
}

// Returns true if this container has this value or false if it does not.
//
func (self *base[V]) Contains(value V) bool {
    lb := self.LowerBound(value)
	if lb < len(self.data) && !self.cmp(value, self.data[lb]) {
	    return true
	} else {
    	return false
    }
}


// This method takes an iterator and will returns true if any of these equivalent values are contained within this
// container.
//
func (self *base[V]) Any(values iter.Seq[V]) bool {
    size := len(self.data)
    for lb, value := range self.traverse(values, self.cmp) {
        if lb < size && !self.cmp(value, self.data[lb]) {
		    return true
		}
    }
    return false
}


// This method takes an iterator and returns true if this container is a superset of these values.
//
func (self *base[V]) All(values iter.Seq[V]) bool {
   size := len(self.data)
   for lb, value := range self.traverse(values, self.cmp) {
        if lb >= size || self.cmp(value, self.data[lb]) {
		    return false
		}
   }
   return true
}


// Returns an index to the first value in the range where the comparison is not less than.
//
func (self *base[V]) LowerBound(value V) int {
    return self.bounds(value, 0, len(self.data) - 1, self.cmp)
}


// Returns an index to the first value in the range where the comparison is greater.
//
func (self *base[V]) UpperBound(value V) int {
    return self.bounds(value, 0, len(self.data) - 1, func(lhs, rhs V) bool { return !self.cmp(rhs, lhs) })
}


// A FlatSet is a sorted associative container of unique values using a comparison function.
//
type FlatSet[V any] struct {
   base[V]
}


// Private method to remove subsequent keys that are repeated.
//
func (self *FlatSet[V]) removeDuplicates() {
    size := len(self.data)
    if size > 1 {
        upto := 1
        for next := 1; next < size; next++ {
            if !self.cmp(self.data[next - 1], self.data[next]) {
                continue
            }
            self.data[upto] = self.data[next]
            upto++
        }
        self.data = append([]V(nil), self.data[:upto]...)
    }
}

// Create a new empty FlatSet.
//
func NewFlatSet[V any](cmp Compare[V]) *FlatSet[V] {
    return &FlatSet[V]{base[V]{cmp: cmp}}
}


// Create a new FlatSet and initialize it with some values. Values that are repeated will be discarded.
//
func InitFlatSet[V any](values []V, cmp Compare[V]) *FlatSet[V] {
    self := &FlatSet[V]{base[V]{cmp: cmp}}
    self.data = append([]V(nil), values...)
    sort.SliceStable(self.data, func(lhs, rhs int) bool {return self.cmp(self.data[lhs], self.data[rhs])})
    self.removeDuplicates()
    return self
}


// Searches for a value within this container, and returns the index for the location of the value or -1 if not found.
//
func (self *FlatSet[V]) Find(value V) int {
    lb := self.LowerBound(value)
	if lb < len(self.data) && !self.cmp(value, self.data[lb]) {
	    return lb
	} else {
    	return -1
    }
}


// Insert a new value. If this value is already contained within this container it will return the index of the existing
// value and false, otherwise it will return the index of the new value and true. If insertion is successful it will
// invalidate any previous indices.
//
func (self *FlatSet[V]) Insert(value V) (int, bool) {
    ub := self.UpperBound(value)
    if ub > 0 && !self.cmp(self.data[ub - 1], value) {
        return ub - 1, false
    } else {
    	self.insert(ub, value)
    	return ub, true
	}
}


// Delete the value at this index from this container.
//
func (self *FlatSet[V]) Erase(index int) {
    self.data = append(self.data[:index], self.data[index+1:]...)
}

// Remove this value if it exists in this container and return true, otherwise return false if it was not found.
//
func (self *FlatSet[V]) Remove(value V) bool {
    index := self.Find(value)
    if index != -1 {
        self.Erase(index)
        return true
    } else {
        return false
    }
}

// Try to replace the value at this index. If the previous value was replaced return true, otherwise return false if
// the new value would result in data being out of sequence. This method allow you to quickly modify a value if you know
// its index, without the need to erase the previous value and insert the new one. This method will not invalidate
// previous indices.
//
func (self *FlatSet[V]) Replace(index int, value V) bool {
    size := len(self.data)
    if index < size {
        if (index > 0 && !self.cmp(self.data[index - 1], value)) ||
            (index < size - 1 && !self.cmp(value, self.data[index + 1])) {
            return false
        }
        self.data[index] = value
        return true
    }
    return false
}

// Append another FlatSet into this one. It is also possible to merge FlatSets that have a different comparison
// function. If a value already exists in this container the new value from the other FlatSet will be discarded to
// maintain order stability. This method is similar but more efficient than Update because it is able to preallocate
// the array. This method updates this container so it will invalidate any previous indices.
//
func (self *FlatSet[V]) Merge(other *FlatSet[V]) {
    if reflect.ValueOf(self.cmp).Pointer() != reflect.ValueOf(other.cmp).Pointer() {
        other = InitFlatSet[V](other.data, self.cmp)
    }
    self.mergeSorted(&other.base)
    self.removeDuplicates()
}


// Insert these values into this container. This method is more flexible but less efficient than Merge because it takes
// a generic iterator of values. If a value already exists in this container the new value will be discarded to maintain
// order stability. This method updates this container so it will invalidate any previous indices.
//
func (self *FlatSet[V]) Update(values iter.Seq[V]) {
    for ub, value := range self.traverse(values, func(lhs, rhs V) bool { return !self.cmp(rhs, lhs) }) {
        if len(self.data) == 0 || (ub > 0 && self.cmp(self.data[ub - 1], value))  {
            self.insert(ub, value)
        }
    }
}

// Return a new FlatSet combining all the values in this container with these other values. If a value already exists in
// the new value will not be included in the resulting FlatSet. This method does not modify this container so it will
// not invalidate previous indices.
//
func (self *FlatSet[V]) Union(values iter.Seq[V]) *FlatSet[V] {
    out := *self
    out.Update(values)
    return &out
}

// Return a new FlatSet containing the common values in this container with these other values. To maintain order
// stability the original values from this container will be returned. This method does not modify this container so it
// will not invalidate previous indices.
//
func (self *FlatSet[V]) Intersection(values iter.Seq[V]) *FlatSet[V] {
    size := len(self.data)
    out := FlatSet[V]{base[V]{cmp: self.cmp}}
    out.data = make([]V, size)

    i := 0
    for lb, value := range self.traverse(values, self.cmp) {
        if lb < size && !self.cmp(value, self.data[lb]) {
            out.data[i] = value
            i++
        }
    }
    out.data = append([]V(nil), out.data[:i]...)
    return &out
}


// Return a new FlatSet containing the values that exist in this container but not in these other values. This method
// does not modify this container so it will not invalidate previous indices.
//
func (self *FlatSet[V]) Difference(values iter.Seq[V]) *FlatSet[V] {
    out := FlatSet[V]{base[V]{cmp: self.cmp}}
    out.data = append([]V(nil), self.data...)

    i := 0
    for lb, value := range out.traverse(values, self.cmp) {
        if lb < len(out.data) && !self.cmp(value, out.data[lb]) {
            size := len(out.data)
            copy(out.data[lb:size-1], out.data[lb+1:size])
            i++
        }
    }
    out.data = append([]V(nil), out.data[:len(self.data) - i]...)
    return &out
}


// A FlatMultiSet is a sorted associative container of values using a comparison function. Unlike a FlatSet, a
// FlatMultiSet allows equivalent values to be stored in the same container and order stability of these values is
// guaranteed.
//
type FlatMultiSet[V any] struct {
   base[V]
}


// Create a new empty FlatMultiSet.
//
func NewFlatMultiSet[V any](cmp Compare[V]) *FlatMultiSet[V] {
    return &FlatMultiSet[V]{base[V]{cmp: cmp}}
}


// Create a new FlatMultiSet and initialize it with some values. The order of equivalent values will be maintained.
//
func InitFlatMultiSet[V any](values []V, cmp Compare[V]) *FlatMultiSet[V] {
    self := &FlatMultiSet[V]{base[V]{cmp: cmp}}
    self.data = append([]V(nil), values...)
    sort.SliceStable(self.data, func(lhs, rhs int) bool {return self.cmp(self.data[lhs], self.data[rhs])})
    return self
}


// Searches for equivalent values within this container, it will return the index of the first value (inclusive) and
// index of the last value exclusive(). If no equivalent value is found this method will return -1, -1.
//
func (self *FlatMultiSet[V]) Find(value V) (int, int) {
    size := len(self.data)
    lb := self.LowerBound(value)
	if lb < size && !self.cmp(value, self.data[lb]) {
	    if lb == size - 1 || (lb < size - 2 && self.cmp(self.data[lb], self.data[lb + 1])) {
	        return lb, lb + 1
	    } else {
	        return lb, self.bounds(value, lb + 1, len(self.data) - 1, func(lhs, rhs V) bool { return !self.cmp(rhs, lhs) })
	    }
	} else {
    	return -1, -1
    }
}


// Insert a new value at the upper bound and return the index of the new value. This method will invalidate any previous
// indices.
//
func (self *FlatMultiSet[V]) Insert(value V) int {
	ub := self.UpperBound(value)
    self.insert(ub, value)
    return ub
}


// Delete values from this index (inclusive) upto this index (exclusive) from this container. If from == -1 this method
// is a no-op in order that you can pass the indices from Find as arguments. This method will invalidate any previous
// indices.
//
func (self *FlatMultiSet[V]) Erase(from, upto int) {
    if from >= 0 {
        self.data = append(self.data[:from], self.data[upto:]...)
    }
}

// Delete any values equivalent to this value and return the number of values that were removed. This method will
// invalidate any previous indices.
//
func (self *FlatMultiSet[V]) Remove(value V) int {
    from, upto := self.Find(value)
    self.Erase(from, upto)
    return upto - from
}


// Try to replace the value at this index. If the previous value was replaced return true, otherwise return false if
// the new value would result in data being out of sequence. This method allow you to quickly modify a value if you know
// its index, without the need to erase the previous value and insert the new one. This method will not invalidate
// previous indices.
//
func (self *FlatMultiSet[V]) Replace(index int, value V) bool {
    size := len(self.data)
    if index < size {
        if (index > 0 && self.cmp(value, self.data[index - 1], )) ||
            (index < size - 1 && self.cmp(self.data[index + 1], value)) {
            return false
        }
        self.data[index] = value
        return true
    }
    return false
}


// Append another FlatMultiSet into this one. It is also possible to merge FlatMultiSets that have a different
// comparison function. Values from the other container will be inserted at the upper bound so equivalent values will be
// ordered after the one in this container other ones. This method is similar but more efficient than Update because it
// is able to preallocate the array. This method will invalidate any previous indices.
//
func (self *FlatMultiSet[V]) Merge(other *FlatMultiSet[V]) {
    if reflect.ValueOf(self.cmp).Pointer() != reflect.ValueOf(other.cmp).Pointer() {
        other = InitFlatMultiSet[V](other.data, self.cmp)
    }
    self.mergeSorted(&other.base)
}


// Insert these values into this container at the upper bound to maintain order stability. This method is more flexible
// but less efficient than Merge because it takes a generic iterator of values. This method updates this container so
// it will invalidate any previous indices.
//
func (self *FlatMultiSet[V]) Update(values iter.Seq[V]) {
    for ub, value := range self.traverse(values, func(lhs, rhs V) bool { return !self.cmp(rhs, lhs) }) {
        self.insert(ub, value)
    }
}
