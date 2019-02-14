package funcdep

import (
	"fmt"
	"sort"
	"strings"
)

// A Relation with a set of functional dependencies.
type Relation struct {
	// Name of the relation.
	Name string

	// Attrs lists all the attributes in the Relation.
	Attrs AttrSet

	// FuncDeps contains all of the functional dependencies over the Relation.
	FuncDeps []*FuncDep
}

func (r *Relation) String() string {
	line := r.Name + "(" + r.Attrs.String() + ")\n\n"
	for _, fd := range r.FuncDeps {
		line += fd.String() + "\n"
	}
	return strings.TrimSpace(line)
}

// RelationFromString parses a relation and optional set of functional dependencies from a string.
func RelationFromString(desc string) (*Relation, error) {
	lines := strings.Split(desc, "\n")
	head := strings.TrimSpace(lines[0])
	pidx := strings.Index(head, "(")
	if pidx == -1 {
		return nil, fmt.Errorf("invalid relation description")
	}
	r := &Relation{}
	r.Name = head[:pidx]

	// trim off parens
	head = head[pidx+1 : len(head)-1]
	for _, s := range strings.Split(head, AttrSep) {
		a := Attr(strings.TrimSpace(s))
		r.Attrs.Add(a)
	}

	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fd, err := FromString(line)
		if err != nil {
			return nil, err
		}
		r.FuncDeps = append(r.FuncDeps, fd)
	}

	var problems AttrSet
	// validate that FDs refer to Attributes in Relation only
	for _, fd := range r.FuncDeps {
		a := fd.Left.Union(fd.Right)
		remAttr := a.Difference(r.Attrs)
		if len(remAttr) != 0 {
			problems.AddAll(remAttr)
		}
	}

	if len(problems) > 0 {
		return nil, fmt.Errorf("relation has %d attributes (%v). FD has %d unknown attributes (%v)",
			len(r.Attrs), r.Attrs, len(problems), problems)
	}

	return r, nil
}

// Closures computes the closure over every Functional Dependency.
func (r *Relation) Closures() []*FuncDep {
	res := make([]*FuncDep, len(r.FuncDeps))
	for i, fd := range r.FuncDeps {
		res[i] = r.Closure(fd)
	}
	return res
}

// Closure computes the closure of the given FD over the functional dependencies on this Relation.
func (r *Relation) Closure(fd *FuncDep) *FuncDep {
	clofd := &FuncDep{}
	clofd.Left.AddAll(fd.Left)
	clofd.Right.AddAll(fd.Left, fd.Right)

	lastN := 0
	n := len(clofd.Right)
	for n != lastN {
		for _, other := range r.FuncDeps {
			if clofd.Right.Contains(other.Left) {
				clofd.Right.AddAll(other.Right)
			}
		}
		lastN = n
		n = len(clofd.Right)
	}
	return clofd
}

// CandidateKeys returns a list of possible keys for the relation.
// This is a somewhat efficient enumeration that relies on at least one
// functional dependency's closure fitting the criteria.
// If this fails to return any candidates, a brute-force search may be necessary.
func (r *Relation) CandidateKeys() []AttrSet {
	var ckeys []AttrSet
	clo := r.Closures()
	for _, cfd := range clo {
		if len(cfd.Right) == len(r.Attrs) {
			ckeys = append(ckeys, cfd.Left)
		}
	}
	return ckeys
}

// CandidateKeysAlt returns a list of possible keys for the relation. This is
// a somewhat efficient enumeration that gets a closure over every functional
// dependency, then augments each with 1-2 missing attributes to fill the
// criteria. If this fails to return any candidates, a brute-force search
// will be necessary.
func (r *Relation) CandidateKeysAlt() []AttrSet {
	var ckeys []AttrSet
	clo := r.Closures()
	for _, cfd := range clo {
		if len(cfd.Right) >= (len(r.Attrs) - 2) {
			diff := r.Attrs.Difference(cfd.Right)
			if len(diff) > 0 {
				cfd.Left.Add(diff[0])
			}
			if len(diff) > 1 {
				cfd.Left.Add(diff[1])
			}
			ckeys = append(ckeys, cfd.Left)
		}
	}
	return ckeys
}

// CandidateKeysBF enumerates all possible keys for the relation using a
// brute-force approach. To reduce number of results, candidate keys that
// contain smaller candidate keys are removed from the result set.
func (r *Relation) CandidateKeysBF() []AttrSet {
	return r.filterContainingKeys(r.enumerateCandidateKeys())
}

func (r *Relation) filterContainingKeys(candidates []AttrSet) []AttrSet {
	// removes candidate keys that fully contain smaller
	// candidate keys recursively.

	if len(candidates) <= 1 {
		return candidates
	}

	// sort candidates so the smallest is at the end
	sort.Slice(candidates, func(i, j int) bool {
		return len(candidates[i]) > len(candidates[j])
	})

	var result []AttrSet
	for len(candidates) > 0 {
		// add the last candidate key to the result
		n := len(candidates) - 1
		x := candidates[n]
		result = append(result, x)

		// to remove containing keys, copy keys that don't
		// contain X forward, then truncate the list
		j := 0
		for i := 0; i < n; i++ {
			if !candidates[i].Contains(x) {
				if j != i {
					candidates[j] = candidates[i]
				}
				j++
			}
		}
		candidates = candidates[:j]
	}
	return result
}

func (r *Relation) enumerateCandidateKeys() []AttrSet {
	var result []AttrSet

	clos := r.Closures()
	hits := make(map[string]struct{})

	check := func(a AttrSet) {
		if _, ok := hits[a.String()]; ok {
			return
		}
		var right AttrSet
		right.AddAll(a)
		last := 0
		n := len(right)
		for n != last {
			for _, c := range clos {
				if right.Contains(c.Left) {
					right.AddAll(c.Right)
				}
			}
			last = n
			n = len(right)
		}
		if len(right) == len(r.Attrs) {
			var x AttrSet
			x.AddAll(a)
			result = append(result, x)
		}
		hits[a.String()] = struct{}{}
	}

	r.recurBF(nil, len(r.Attrs), check)
	return result
}

func (r *Relation) recurBF(x AttrSet, nremain int, check func(AttrSet)) {
	var z AttrSet
	z.AddAll(x)
	for _, a1 := range r.Attrs {
		if z.Add(a1) {
			check(z)
			if nremain > 1 {
				r.recurBF(z, nremain-1, check)
			}

			z.Remove(a1)
		}
	}
}
