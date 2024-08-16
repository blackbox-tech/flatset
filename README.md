# flatset

This is a golang module that provides the sorted associative containers 'FlatSet' and 'FlatMultiSet' (similar to C++ 
std::flat_set and std::flat_multiset).

These containers store their data in contiguous memory instead of using a traditional binary-tree structure or 
hash-table.

This has many advantages over standard associative containers:
  * Faster lookup
  * Much faster iteration
  * Random-access to the underlying array
  * Less memory consumption
  * Improved cache performance

The disadvantages are:
  * previous indices are invalidated following insertion or deletion.
  * insertion can be slower, especially when inserting near the beginning of a large collection

The module implements an algorithm that can iterate over other types of containers, traversing the flatset so that it 
can be accessed and updated more efficiently than processing each value individually (making use of golang's 'range over 
functions' to do this).

### Simple example
```go
import (
    "fmt"
    "github.com/blackbox-tech/flatset"
)

func lessInt(lhs, rhs int) bool { 
    return lhs < rhs 
}

func main() {

    // insertion
    a := flatset.NewFlatSet[int](lessInt)
    a.Insert(8)
    a.Insert(5)
    a.Insert(10)
    a.Insert(6)

    // find and erase
    idx := a.Find(6)
    fmt.Printf("a contains %d at index %d\n", a.At(idx), idx)
    a.Erase(idx)
    fmt.Printf("a erased value at index %d\n", idx)
    fmt.Printf("a contains 6 == %t\n", a.Contains(6))

    // array intersection
    b := flatset.InitFlatSet[int]([]int{6, 8, 10}, lessInt)
    c := a.Intersection(b.All())
    for value := range c.All() {
        fmt.Printf("%d is in a and b\n", value)
    }
}
```
### Output: 
```
a contains 6 at index 1
a erased value at index 1
a contains 6 == false
8 is in a and b
10 is in a and b
```

For more information see the [API Reference](./REFERENCE.md)

### License

This project is licensed under the terms of the [MIT license](./LICENSE)
