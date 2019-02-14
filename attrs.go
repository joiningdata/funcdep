package funcdep

import (
	"sort"
	"strings"
)

// AttrSep is the attribute separator.
var AttrSep = ","

// Attr represents an attribute in a functional dependency.
type Attr string

// AttrSet is a set of attributes in a functional dependency.
type AttrSet []Attr

// String representation of the attribute set (joined by commas).
func (s AttrSet) String() string {
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
	sb := strings.Builder{}
	for i, a := range s {
		if i > 0 {
			sb.WriteString(AttrSep)
		}
		sb.WriteString(string(a))
	}
	return sb.String()
}

// Contains returns true if this AttrSet contains all elements of other.
// (e.g. other is a subset of this)
func (s AttrSet) Contains(other AttrSet) bool {
	chk := make(map[Attr]struct{})
	for _, x := range s {
		chk[x] = struct{}{}
	}
	for _, x := range other {
		if _, ok := chk[x]; !ok {
			return false
		}
	}
	return true
}

// Add an attribute to this attribute set if not already present.
// Returns true if the element was added.
func (s *AttrSet) Add(a Attr) bool {
	for _, x := range *s {
		if x == a {
			return false
		}
	}
	*s = append(*s, a)
	return true
}

// AddAll adds all attributes in the other sets if not already present.
// Somewhat more efficient for large additions than Add().
func (s *AttrSet) AddAll(others ...AttrSet) {
	chk := make(map[Attr]struct{})
	for _, x := range *s {
		chk[x] = struct{}{}
	}
	for _, other := range others {
		for _, x := range other {
			if _, ok := chk[x]; !ok {
				*s = append(*s, x)
				chk[x] = struct{}{}
			}
		}
	}
}

// Remove an attribute from this attribute set if it is present.
// Returns true if the element was removed.
func (s *AttrSet) Remove(a Attr) bool {
	for i, x := range *s {
		if x == a {
			xs := []Attr(*s)
			if len(xs) == 1 {
				*s = AttrSet(xs[:0])
				return true
			}
			j := len(xs) - 1
			if i != j {
				// swap with the last element
				xs[i], xs[j] = xs[j], xs[i]
			}
			// remove the last element
			*s = AttrSet(xs[:j])
			return true
		}
	}
	return false
}

// Union of this and the other attribute sets, returned as a new AttrSet.
func (s AttrSet) Union(others ...AttrSet) AttrSet {
	// not the most efficient...
	var res AttrSet
	for _, x := range s {
		res.Add(x)
	}
	for _, other := range others {
		for _, x := range other {
			res.Add(x)
		}
	}
	return res
}

// Intersection of this and the other attribute sets, returned as a new AttrSet.
func (s AttrSet) Intersection(others ...AttrSet) AttrSet {
	// not the most efficient...
	inter := make(map[Attr]int)
	for _, x := range s {
		inter[x] = 1
	}
	var res AttrSet
	for _, other := range others {
		for _, x := range other {
			if n, ok := inter[x]; ok {
				inter[x] = n + 1
			}
		}
	}
	n := len(others) + 1
	for x, c := range inter {
		if c == n {
			res.Add(x)
		}
	}
	return res
}

// Difference removes all the elements of the other attribute sets from this
// AttrSet and returns a new AttrSet with the remaining attributes.
func (s AttrSet) Difference(others ...AttrSet) AttrSet {
	// not the most efficient...
	inter := make(map[Attr]struct{})
	for _, x := range s {
		inter[x] = struct{}{}
	}
	var res AttrSet
	for _, other := range others {
		for _, x := range other {
			delete(inter, x)
			if len(inter) == 0 {
				return res
			}
		}
	}
	for x := range inter {
		res.Add(x)
	}
	return res
}
