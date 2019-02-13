package funcdep

// Augment this functional dependency with Attribute a and return a new FuncDep.
//    if A->B and C is an attribute, then AC->BC
//
// Always returns true.
func (fd *FuncDep) Augment(a Attr) (*FuncDep, bool) {
	res := &FuncDep{}
	res.Left.AddAll(fd.Left)
	res.Right.AddAll(fd.Right)
	res.Left.Add(a)
	res.Right.Add(a)
	return res, true
}

// TransitiveWith uses Transitivity to combine this and the other functional
// dependencies, returning a new FuncDep with the intermediate attributes removed.
//    if A->B and B->C then A->C
// This implementation also does pseudo-transitivity:
//    if A->BX and B->C then A->C  (e.g. X is ignored/dropped)
//
// Returns false if neither functional dependency's right side contains a
// superset of the other's left side.
func (fd *FuncDep) TransitiveWith(other *FuncDep) (*FuncDep, bool) {
	left := fd.Left
	right := fd.Right

	// first determine order (A->B B->C) or (B->C A->B)

	right2Left := fd.Right.Intersection(other.Left)
	left2Right := other.Right.Intersection(fd.Left)
	if len(right2Left) == len(other.Left) {
		// A->B1, B2->C is good!
		// B2 is a subset of B1, so A->C ok
		right = other.Right
	} else if len(left2Right) == len(fd.Left) {
		// A1->B, C->A2 is good but needs swapped!
		// A1 is a subset of A2, so C->B ok
		left = other.Left
	} else {
		// cannot apply transitivity!
		return nil, false
	}

	res := &FuncDep{}
	res.Left.AddAll(left)
	res.Right.AddAll(right)
	return res, true
}

// Decompose this functional dependency by enumerating new functional dependencies
// for each attribute independently on the right side.
//    if A->XYZ then A->X, A->Y, and A->Z
//
// Returns false if there is only 1 attribute on the right side.
func (fd *FuncDep) Decompose() ([]*FuncDep, bool) {
	if len(fd.Right) == 1 {
		return nil, false
	}
	var results []*FuncDep
	for _, a := range fd.Right {
		res := &FuncDep{}
		res.Left.AddAll(fd.Left)
		res.Right.Add(a)

		results = append(results, res)
	}
	return results, true
}

// Compose this functional dependency with the other and return a new functional dependency.
//    if A->B and X->Y then AX->BY
//
// Always returns true.
func (fd *FuncDep) Compose(other *FuncDep) (*FuncDep, bool) {
	res := &FuncDep{}
	res.Left.AddAll(fd.Left, other.Left)
	res.Right.AddAll(fd.Right, other.Right)
	return res, true
}

// Union this functional dependency with the other and return a new functional dependency.
//    if A->B and A->C then A->BC
//
// Returns false if the left sides do not match.
func (fd *FuncDep) Union(other *FuncDep) (*FuncDep, bool) {
	chk := make(map[Attr]struct{})
	res := &FuncDep{}
	for _, x := range fd.Left {
		res.Left.Add(x)
		chk[x] = struct{}{}
	}
	for _, x := range other.Left {
		if _, ok := chk[x]; !ok {
			return nil, false
		}
	}
	res.Right.AddAll(fd.Right, other.Right)
	return res, true
}
