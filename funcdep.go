// Package funcdep has various tools and methods for working with functional dependencies.
package funcdep

import (
	"fmt"
	"regexp"
	"strings"
)

// FuncDep represents a functional dependency of the form:
//    Left --> Right
type FuncDep struct {
	Left  AttrSet
	Right AttrSet
}

// String representation of the functional dependency (joined by an ASCII arrow).
func (fd *FuncDep) String() string {
	return fd.Left.String() + " --> " + fd.Right.String()
}

// accepts multiple forms of left->right arrows:
//   > --> ---> >> -->>
//   ~~> ~> ==> ===>>
//   → ⇒ ⇾  (Unicode arrows)
var cutArrows = regexp.MustCompile("[-=~]*[>→⇒⇾]+")

// FromString converts a text/string description of a functional dependency into
// a parsed FuncDep structure. It accepts multiple forms of arrows in the
// representation (as long as they point to the right).
func FromString(fdesc string) (*FuncDep, error) {
	parts := cutArrows.Split(fdesc, -1)
	if len(parts) == 1 {
		return nil, fmt.Errorf("no arrow found in functional dependency")
	}
	if len(parts) != 2 {
		return nil, fmt.Errorf("too many arrows in functional dependency")
	}
	fd := &FuncDep{}
	for _, s := range strings.Split(parts[0], AttrSep) {
		a := Attr(strings.TrimSpace(s))
		fd.Left = append(fd.Left, a)
	}
	for _, s := range strings.Split(parts[1], AttrSep) {
		a := Attr(strings.TrimSpace(s))
		fd.Right = append(fd.Right, a)
	}
	return fd, nil
}
