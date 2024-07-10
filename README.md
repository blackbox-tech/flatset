# flatset

Package flatset provides the sorted associative containers 'FlatSet' and
'FlatMultiSet' that store data in continuous memory instead of a binary tree
(similar to C++ std::flat_set and std::flat_multiset).

Both the FlatSet and FlatMultiSet implement stable ordering, but the previous
indices for values are invalidated following an method that modifies the
container.

Compared to a binary tree, flatsets are considerably faster to read as they
avoid the cache misses from pointer referencing. On modern CPUs it is also
surprisingly fast to insert into a flatset, especially if the collection is
small, or if you frequently update the end of the flatset. For larger
collections you can optimize write operations by using a flatset of pointers to
structures instead of values, although you must be careful not to modify the
data inside the flatset as it might affect how the data is sorted.

This package requires 'range over functions' so for golang < 1.23 you must 
build your application with 'GOEXPERIMENT=rangefunc'.
https://go.dev/wiki/RangefuncExperiment

### License

This project is licensed under the terms of the [MIT license](./LICENSE)

___

# API Reference

## Compare

```go
type Compare[V any] func(a, b V) bool
```

This is the interface for the comparison function that is passed to the FlatSet
and FlatMultiSet which defines how the data will be sorted. For example, to sort
the data in ascending order the comparison function would implement less than.
___

## FlatSet

```go
type FlatSet[V any] struct {
}
```

A FlatSet is a sorted associative container of unique values using a comparison
function.

#### func  InitFlatSet

```go
func InitFlatSet[V any](values []V, cmp Compare[V]) *FlatSet[V]
```
Create a new FlatSet and initialize it with some values. Values that are
repeated will be discarded.

#### func  NewFlatSet

```go
func NewFlatSet[V any](cmp Compare[V]) *FlatSet[V]
```
Create a new empty FlatSet.

### Methods

#### func (*FlatSet) At

```go
func (self *FlatSet) At(index int) V
```
Returns a copy of the value at the given index.

#### func (*FlatSet) Size

```go
func (self *FlatSet) Size() int
```
Returns the number of values stored in this container.

#### func (*FlatSet) Values

```go
func (self *FlatSet) Values() iter.Seq[V]
```
Returns an iterator that returns a copy of each value in order.

#### func (*FlatSet) Reversed

```go
func (self *FlatSet) Reversed() iter.Seq[V]
```
Returns an iterator that iterates in reverse order returning a copy of each
value.

#### func (*FlatSet) Contains

```go
func (self *FlatSet) Contains(value V) bool
```
Returns true if this container has this value or false if it does not.

#### func (*FlatSet) Any

```go
func (self *FlatSet) Any(values iter.Seq[V]) bool
```
This method takes an iterator and will returns true if any of these equivalent 
values are contained within this container.

#### func (*FlatSet) All

```go
func (self *FlatSet) All(values iter.Seq[V]) bool
```
This method takes an iterator and returns true if this container is a superset
of these values.

#### func (*FlatSet) LowerBound

```go
func (self *FlatSet) LowerBound(value V) int
```
Returns an index to the first value in the range where the comparison is not
less than.

#### func (*FlatSet) UpperBound

```go
func (self *FlatSet) UpperBound(value V) int
```
Returns an index to the first value in the range where the comparison is
greater.

#### func (*FlatSet[V]) Find

```go
func (self *FlatSet[V]) Find(value V) int
```
Searches for a value within this container, and returns the index for the
location of the value or -1 if not found.

#### func (*FlatSet[V]) Insert

```go
func (self *FlatSet[V]) Insert(value V) (int, bool)
```
Insert a new value. If this value is already contained within this container it
will return the index of the existing value and false, otherwise it will return
the index of the new value and true. If insertion is successful it will
invalidate any previous indices.

#### func (*FlatSet[V]) Erase

```go
func (self *FlatSet[V]) Erase(index int)
```
Delete the value at this index from this container.

#### func (*FlatSet[V]) Remove

```go
func (self *FlatSet[V]) Remove(value V) bool
```
Remove this value if it exists in this container and return true, otherwise
return false if it was not found.

#### func (*FlatSet[V]) Replace

```go
func (self *FlatSet[V]) Replace(index int, value V) bool
```
Try to replace the value at this index. If the previous value was replaced
return true, otherwise return false if the new value would result in data being
out of sequence. This method allow you to quickly modify a value if you know its
index, without the need to erase the previous value and insert the new one. This
method will not invalidate previous indices.

#### func (*FlatSet[V]) Merge

```go
func (self *FlatSet[V]) Merge(other *FlatSet[V])
```
Append another FlatSet into this one. It is also possible to merge FlatSets that
have a different comparison function. If a value already exists in this
container the new value from the other FlatSet will be discarded to maintain
order stability. This method is similar but more efficient than Update because
it is able to preallocate the array. This method updates this container so it
will invalidate any previous indices.

#### func (*FlatSet[V]) Update

```go
func (self *FlatSet[V]) Update(values iter.Seq[V])
```
Insert these values into this container. This method is more flexible but less
efficient than Merge because it takes a generic iterator of values. If a value
already exists in this container the new value will be discarded to maintain
order stability. This method updates this container so it will invalidate any
previous indices.

#### func (*FlatSet[V]) Union

```go
func (self *FlatSet[V]) Union(values iter.Seq[V]) *FlatSet[V]
```
Return a new FlatSet combining all the values in this container with these other
values. If a value already exists in the new value will not be included in the
resulting FlatSet. This method does not modify this container so it will not
invalidate previous indices.

#### func (*FlatSet[V]) Intersection

```go
func (self *FlatSet[V]) Intersection(values iter.Seq[V]) *FlatSet[V]
```
Return a new FlatSet containing the common values in this container with these
other values. To maintain order stability the original values from this
container will be returned. This method does not modify this container so it
will not invalidate previous indices.

#### func (*FlatSet[V]) Difference

```go
func (self *FlatSet[V]) Difference(values iter.Seq[V]) *FlatSet[V]
```
Return a new FlatSet containing the values that exist in this container but not
in these other values. This method does not modify this container so it will not
invalidate previous indices.

___

## FlatMultiSet

```go
type FlatMultiSet[V any] struct {
}
```

A FlatMultiSet is a sorted associative container of values using a comparison
function. Unlike a FlatSet, a FlatMultiSet allows equivalent values to be stored
in the same container and order stability of these values is guaranteed.

#### func  InitFlatMultiSet

```go
func InitFlatMultiSet[V any](values []V, cmp Compare[V]) *FlatMultiSet[V]
```
Create a new FlatMultiSet and initialize it with some values. The order of
equivalent values will be maintained.

#### func  NewFlatMultiSet

```go
func NewFlatMultiSet[V any](cmp Compare[V]) *FlatMultiSet[V]
```
Create a new empty FlatMultiSet.

### Methods

#### func (*FlatMultiSet) At

```go
func (self *FlatMultiSet) At(index int) V
```
Returns a copy of the value at the given index.

#### func (*FlatMultiSet) Size

```go
func (self *FlatMultiSet) Size() int
```
Returns the number of values stored in this container.

#### func (*FlatMultiSet) Values

```go
func (self *FlatMultiSet) Values() iter.Seq[V]
```
Returns an iterator that returns a copy of each value in order.

#### func (*FlatMultiSet) Reversed

```go
func (self *FlatMultiSet) Reversed() iter.Seq[V]
```
Returns an iterator that iterates in reverse order returning a copy of each
value.

#### func (*FlatMultiSet) Contains

```go
func (self *FlatMultiSet) Contains(value V) bool
```
Returns true if this container has this value or false if it does not.

#### func (*FlatMultiSet) Any

```go
func (self *FlatMultiSet) Any(values iter.Seq[V]) bool
```
This method takes an iterator and will returns true if any of these equivalent 
values are contained within this container.

#### func (*FlatMultiSet) All

```go
func (self *FlatMultiSet) All(values iter.Seq[V]) bool
```
This method takes an iterator and returns true if this container is a superset
of these values.

#### func (*FlatMultiSet) LowerBound

```go
func (self *FlatMultiSet) LowerBound(value V) int
```
Returns an index to the first value in the range where the comparison is not
less than.

#### func (*FlatMultiSet) UpperBound

```go
func (self *FlatMultiSet) UpperBound(value V) int
```
Returns an index to the first value in the range where the comparison is
greater.

#### func (*FlatMultiSet[V]) Find

```go
func (self *FlatMultiSet[V]) Find(value V) (int, int)
```
Searches for equivalent values within this container, it will return the index
of the first value (inclusive) and index of the last value exclusive(). If no
equivalent value is found this method will return -1, -1.

#### func (*FlatMultiSet[V]) Insert

```go
func (self *FlatMultiSet[V]) Insert(value V) int
```
Insert a new value at the upper bound and return the index of the new value.
This method will invalidate any previous indices.

#### func (*FlatMultiSet[V]) Erase

```go
func (self *FlatMultiSet[V]) Erase(from, upto int)
```
Delete values from this index (inclusive) upto this index (exclusive) from this
container. If from == -1 this method is a no-op in order that you can pass the
indices from Find as arguments. This method will invalidate any previous
indices.

#### func (*FlatMultiSet[V]) Remove

```go
func (self *FlatMultiSet[V]) Remove(value V) int
```
Delete any values equivalent to this value and return the number of values that
were removed. This method will invalidate any previous indices.

#### func (*FlatMultiSet[V]) Replace

```go
func (self *FlatMultiSet[V]) Replace(index int, value V) bool
```
Try to replace the value at this index. If the previous value was replaced
return true, otherwise return false if the new value would result in data being
out of sequence. This method allow you to quickly modify a value if you know its
index, without the need to erase the previous value and insert the new one. This
method will not invalidate previous indices.

#### func (*FlatMultiSet[V]) Merge

```go
func (self *FlatMultiSet[V]) Merge(other *FlatMultiSet[V])
```
Append another FlatMultiSet into this one. It is also possible to merge
FlatMultiSets that have a different comparison function. Values from the other
container will be inserted at the upper bound so equivalent values will be
ordered after the one in this container other ones. This method is similar but
more efficient than Update because it is able to preallocate the array. This
method will invalidate any previous indices.

#### func (*FlatMultiSet[V]) Update

```go
func (self *FlatMultiSet[V]) Update(values iter.Seq[V])
```
Insert these values into this container at the upper bound to maintain order
stability. This method is more flexible but less efficient than Merge because it
takes a generic iterator of values. This method updates this container so it
will invalidate any previous indices.
