package funcdep

import (
	"fmt"
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
		r.Attrs = append(r.Attrs, a)
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
