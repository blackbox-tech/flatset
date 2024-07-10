package flatset

import (
    "iter"
    "math/rand"
    "strings"
    "testing"
)


// There should be a std iter.Values function in golang >=1.23
func values[T any](values []T) iter.Seq[T] {
    return func(yield func(T) bool) {
        for _, value := range values {
            if !yield(value) {
                break
            }
        }
    }
}


func lessInt(lhs, rhs int) bool { return lhs < rhs }
func greaterInt(lhs, rhs int) bool { return lhs > rhs }


// Test the LowerBound and UpperBound methods for the FlatSet.
//
func TestBoundsUniq(t *testing.T) {
    fs := InitFlatSet[int]([]int {2, 4}, lessInt)

    for value, expected := range map[int]int {1: 0, 2: 0, 3: 1, 4: 1, 5: 2} {
        actual := fs.LowerBound(value)
        if actual != expected {
            t.Errorf("FlatSet.LowerBound(%d): expected(%d), actual(%d)", value, expected, actual)
        }
    }

    for value, expected := range map[int]int {1: 0, 2: 1, 3: 1, 4: 2, 5: 2} {
        actual := fs.UpperBound(value)
        if actual != expected {
            t.Errorf("FlatSet.UpperBound(%d): expected(%d), actual(%d)", value, expected, actual)
        }
    }
}


// Test the LowerBound and UpperBound methods for the FlatMultiSet.
//
func TestBoundMulti(t *testing.T) {
    fs := InitFlatMultiSet[int]([]int {2, 2, 4}, lessInt)

    for value, expected := range map[int]int {1: 0, 2: 0, 3: 2, 4: 2, 5: 3} {
        actual := fs.LowerBound(value)
        if actual != expected {
            t.Errorf("FlatMultiSet.LowerBound(%d): expected(%d), actual(%d)", value, expected, actual)
        }
    }

    for value, expected := range map[int]int {1: 0, 2: 2, 3: 2, 4: 3, 5: 3} {
        actual := fs.UpperBound(value)
        if actual != expected {
            t.Errorf("FlatMultiSet.UpperBound(%d): expected(%d), actual(%d)", value, expected, actual)
        }
    }
}


// Test the Insert/Find/Replace methods for the FlatSet.
//
func TestInsertFindReplaceUniq(t *testing.T) {
    type testData struct {
        value int
        index int
        success bool
    }
    fs := NewFlatSet[int](lessInt)

    for _, test := range []testData {{2, 0, true}, {3, 1, true}, {2, 0, false}, {1, 0, true}, {5, 3, true}} {
        index, success := fs.Insert(test.value)
        if index != test.index || success != test.success {
            t.Errorf("FlatSet.Insert(%d): expected(%d, %t), actual(%d, %t)", test.value, test.index, test.success, index,
                     success)
        }
    }

    for value, expected := range map[int]int {0: -1, 1: 0, 3: 2, 4: -1, 5: 3} {
        index := fs.Find(value)
        if index != expected {
            t.Errorf("FlatSet.Find(%d): expected(%d), actual(%d)", value, expected, index)
        }
    }

    for _, test := range []testData {{0, 0, true}, {3, 1, false}, {3, 2, true}, {4, 3, true}, {4, 2, false},
                                     {6, 3, true}} {
        success := fs.Replace(test.index, test.value)
        if success != test.success {
            t.Errorf("FlatSet.Replace(%d, %d): expected(%t), actual(%t)", test.index, test.value, test.success, success)
        }
    }
}


// Test the Insert/Find/Replace methods for the FlatMultiSet.
//
func TestInsertFindReplaceMulti(t *testing.T) {
    type testData struct {
        value int
        index int
        success bool
    }

    fs := NewFlatMultiSet[int](lessInt)

    expected := []int {0, 0, 2, 3, 1, 5}
    for i, value := range []int {3, 1, 4, 5, 1, 5} {
        index := fs.Insert(value)
        if index != expected[i] {
            t.Errorf("FlatMultiSet.Insert(%+v): expected_index(%d), actual(%d)", value, expected[i], index)
        }
    }

    for value, expected := range map[int][2]int {
        0: {-1, -1}, 1: {0, 2}, 2: {-1, -1}, 3: {2, 3}, 5: {4, 6}, 6: {-1, -1}} {
        from, upto := fs.Find(value)
        if from != expected[0] || upto != expected[1] {
            t.Errorf("FlatMultiSet.Find(%+v): expected(%d, %d), actual(%d, %d)", value, expected[0], expected[1], from, upto)
        }
    }

    for _, test := range []testData {{0, 0, true},  {2, 1, true}, {1, 2, false}, {3, 3, true}, {6, 4, false},
        {6, 5, true}, {4, 5, false}} {
        success := fs.Replace(test.index, test.value)
        if success != test.success {
            t.Errorf("FlatMultiSet.Replace(%d, %d): expected(%t), actual(%t)", test.index, test.value, test.success, success)
        }
    }
}


type stableData struct {
    value int
    order int
}


func stableCompare(lhs, rhs stableData) bool {
    return lhs.value < rhs.value
}


func stableCompare2(lhs, rhs stableData) bool {
    return (uint32(lhs.value) << 30) < (uint32(rhs.value) << 30)
}


var stableInit = []stableData {{4, 0}, {2, 2}, {4, 3}, {2, 4}, {2, 5}, {1, 6}}
var stableUpdate = []stableData {{4, 7}, {3, 8}, {5, 9}, {2, 10}}


// Test the new values do not replace existing values in a FlatSet.
//
func TestStableUniq(t *testing.T) {
    fs := InitFlatSet[stableData](stableInit, stableCompare)
    fs2 := fs
    fs3 := fs

    expected := []stableData {{1, 6}, {2, 2}, {4, 0}}
    i := 0
    for actual := range fs.Values()  {
        if expected[i] != actual {
            t.Errorf("InitFlatSet not stable expected(%+v), actual(%+v)", expected[i], actual)
        }
        i++
    }

    fs.Update(values(stableUpdate))

    expected = []stableData {{1, 6}, {2, 2}, {3, 8}, {4, 0}, {5, 9}}
    i = 0
    for actual := range fs.Values()  {
        if expected[i] != actual {
            t.Errorf("FlatSet.Updated not stable expected(%+v), actual(%+v)", expected[i], actual)
        }
        i++
    }

    fs2MergeSorted := InitFlatSet[stableData](stableUpdate, stableCompare)
    fs2.Merge(fs2MergeSorted)
    fs3MergeUnsorted := InitFlatSet[stableData](stableUpdate, stableCompare2)
    fs3.Merge(fs3MergeUnsorted)
    if fs2 != fs {
        t.Errorf("FlatSet.Merge() sorted is not stable")
    } else if fs3 != fs {
        t.Errorf("FlatSet.Merge() unsorted is not stable")
    }
}


// Test the order stability of a FlatMultiSet.
//
func TestStableMulti(t *testing.T) {
    fs := InitFlatMultiSet[stableData](stableInit, stableCompare)
    fs2 := fs
    fs3 := fs

    expected := []stableData {{1, 6}, {2, 2}, {2, 4}, {2, 5}, {4, 0}, {4, 3}}
    i := 0
    for actual := range fs.Values()  {
        if expected[i] != actual {
            t.Errorf("InitFlatMultiSet not stable expected(%+v), actual(%+v)", expected[i], actual)
        }
        i++
    }

    fs.Update(values(stableUpdate))

    expected = []stableData {{1, 6}, {2, 2}, {2, 4}, {2, 5}, {2, 10}, {3, 8}, {4, 0}, {4, 3}, {4, 7}, {5, 9}}
    i = 0
    for actual := range fs.Values()  {
        if expected[i] != actual {
            t.Errorf("FlatMultiSet.Update() not stable expected(%+v), actual(%+v)", expected[i], actual)
        }
        i++
    }

    fs2MergeSorted := InitFlatMultiSet[stableData](stableUpdate, stableCompare)
    fs2.Merge(fs2MergeSorted)
    fs3MergeUnsorted := InitFlatMultiSet[stableData](stableUpdate, stableCompare2)
    fs3.Merge(fs3MergeUnsorted)
    if fs2 != fs {
        t.Errorf("FlatMultiSet.Merge() sorted is not stable")
    } else if fs3 != fs {
        t.Errorf("FlatMultiSet.Merge() unsorted is not stable")
    }
}


// Test the Any/All/Union/Intersection/Difference methods of a FlatSet.
//
func TestSetOperations(t *testing.T) {
    fs := InitFlatSet[int]([]int {2, 4, 5}, lessInt)
    one := []int {1}
    has := []int {2, 4}
    other := []int {2, 3, 5, 6}

    if fs.Any(values(one)) || !fs.Any(values(has)) || !fs.Any(values(other)) {
        t.Errorf("FlatSet.Any() failed")
    }

    if fs.All(values(one)) || !fs.All(values(has)) || fs.All(values(other)) {
        t.Errorf("FlatSet.All() failed")
    }

    fs2 := fs.Union(values(other))
    expected := []int {2, 3, 4, 5, 6}
    i := 0
    for value := range fs2.Values() {
        if value != expected[i] {
            t.Errorf("FlatSet.Union() unexpected value")
        }
        i++
    }

    fs2 = fs.Intersection(values(other))
    expected = []int {2, 5}
    i = 0
    for value := range fs2.Values() {
        if value != expected[i] {
            t.Errorf("FlatSet.Intersection() unexpected value")
        }
        i++
    }

    fs2 = fs.Difference(values(other))
    expected = []int {4}
    i = 0
    for value := range fs2.Values() {
        if value != expected[i] {
            t.Errorf("FlatSet.Difference() unexpected value")
        }
        i++
    }
}


type person struct {
    age int
    name string
}


func comparePeople(lhs, rhs *person) bool { // oldest to youngest, then alphabetically
    if lhs.age > rhs.age {
        return true
    } else if lhs.age < rhs.age {
        return false
    } else {
        return strings.Compare(lhs.name, rhs.name) < 0
    }
}

// Test a FlatSet of pointer to structures instead of values.
//
func TestPointerSet(t *testing.T) {
    originalMembers := []*person {
        &person{name: "Mick", age: 80},
        &person{name: "Keith", age: 80},
        &person{name: "Brian", age: 81},
        &person{name: "Bill", age: 87},
        &person{name: "Charlie", age: 82},
    }

    otherMembers := []*person {
        &person{name: "Ian", age: 85},
        &person{name: "Mick", age: 75},
    }

    fs := InitFlatSet[*person](originalMembers, comparePeople)
    fs.Remove(originalMembers[2])
    fs.Update(values(otherMembers))
    fs.Insert(&person{name: "Ronnie", age: 77})

    fs.Remove(originalMembers[3])
    fs = fs.Difference(values(otherMembers))
    fs.Remove(originalMembers[4])

    expected := [] person { {80, "Keith"}, {80, "Mick"}, {77, "Ronnie"}}
    i := 0
    for value := range fs.Values() {
        for *value != expected[i] {
            t.Errorf("TestPointerSet unexpected value")
        }
        i++
    }
}

//
// Benchmarks
//

func randInt(min, max, size int) []int {
    out := make([]int, size)
    rng := max - min
    for i := 0; i < size; i++ {
        out[i] = rand.Intn(rng) + min
    }
    return out
}

var bmInit = InitFlatSet(randInt(0, 1000000, 100000), lessInt)
var bmRandomInts = randInt(0, 1000000, 10000)
var bmInsertForward = InitFlatSet(bmRandomInts, lessInt)
var bmInsertReversed = InitFlatSet(bmRandomInts, greaterInt)


// Insert each element which will insert using O(log n) complexity.
//
func BenchmarkInsertEach(b *testing.B) {
    out := bmInit
    for value := range values(bmRandomInts) {
        out.Insert(value)
    }
}


// The internal traverse algo typically inserts items in a random order similar to O(log n) complexity insertion.
//
func BenchmarkUpdateRandom(b *testing.B) {
    out := bmInit
    out.Update(values(bmRandomInts))
}


// The internal traverse algo also inserts sorted items faster than O(log n) complexity insertion.
//
func BenchmarkUpdateForward(b *testing.B) {
    out := bmInit
    out.Update(bmInsertForward.Values())
}


// The internal traverse algo also inserts items sorted in reverse order faster than O(log n) complexity insertion.
//
func BenchmarkUpdateReverse(b *testing.B) {
    out := bmInit
    out.Update(bmInsertReversed.Reversed())
}


// Merge with the same compare function is faster than Update as it's a simple sorted merge into a pre-allocated slice.
//
func BenchmarkMergeForward(b *testing.B) {
    out := bmInit
    out.Merge(bmInsertForward)
}


// Merge with a different compare function is slower because it has to sort the new data but it can still pre-allocate.
//
func BenchmarkMergeReverse(b *testing.B) {
    out := bmInit
    out.Merge(bmInsertReversed)
}
