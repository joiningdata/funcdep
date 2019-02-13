package funcdep

// A Set of functional dependencies over a relation.
type Set struct {
	// Attrs lists all the attributes in the set of functional dependencies.
	Attrs AttrSet

	// FuncDeps contains all of the functional dependencies in this set.
	FuncDeps []FuncDep
}
