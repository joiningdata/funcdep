package funcdep

import "strings"

// Attr represents an attribute in a functional dependency.
type Attr string

// AttrSet is a set of attributes in a functional dependency.
type AttrSet []Attr

// String representation of the attribute set (joined by commas).
func (s AttrSet) String() string {
	sb := strings.Builder{}
	for i, a := range s {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(string(a))
	}
	return sb.String()
}

// Add an attribute to the attribute set if not already present.
func (s *AttrSet) Add(a Attr) {
	for _, x := range *s {
		if x == a {
			return
		}
	}
	*s = append(*s, a)
}

// AddAll efficiently adds all attributes in the other sets if not already present.
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

// Remove an attribute from the attribute set if it is present.
func (s *AttrSet) Remove(a Attr) {
	for i, x := range *s {
		if x == a {
			xs := []Attr(*s)
			if len(xs) == 1 {
				*s = AttrSet(xs[:0])
				return
			}
			j := len(xs) - 1
			if i != j {
				// swap with the last element
				xs[i], xs[j] = xs[j], xs[i]
			}
			// remove the last element
			*s = AttrSet(xs[:j])
			return
		}
	}
}

// Union this and the other attribute sets and return a new AttrSet.
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

// Intersection of this and the other attribute sets and return a new AttrSet.
func (s AttrSet) Intersection(others ...AttrSet) AttrSet {
	// not the most efficient...
	inter := make(map[Attr]struct{})
	for _, x := range s {
		inter[x] = struct{}{}
	}
	var res AttrSet
	for _, other := range others {
		for _, x := range other {
			if _, ok := inter[x]; ok {
				res.Add(x)
			} else {
				inter[x] = struct{}{}
			}
		}
	}
	return res
}

// Difference removes all the elements of the other attribute sets from this
// AttrSet and return a new AttrSet of the remaining attributes.
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
