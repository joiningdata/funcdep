// Command data2fd reads in a tabular data file and generates functional dependencies based on observations.
package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joiningdata/funcdep"
)

// DataSet represents data loaded from tabular files amd used to generate
// functional dependencies.
type DataSet struct {
	skiplist map[int]string
	header   []string
	data     [][]string

	rel *funcdep.Relation
}

// ReadData loads a DataSet, tracking the header along with the rows of data.
// Supports both CSV and tab-delimited data files with a single-line header.
func ReadData(filename string) (*DataSet, error) {
	// TODO: support gzip transparently
	// TODO: random sampling for large data files

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ext := filepath.Ext(filename)
	relname := strings.TrimSuffix(filepath.Base(filename), ext)
	ds := &DataSet{
		skiplist: make(map[int]string),
		rel: &funcdep.Relation{
			Name: relname,
		},
	}

	if strings.ToLower(ext) == ".csv" {
		rdr := csv.NewReader(f)
		data, err := rdr.ReadAll()
		if err != nil {
			return nil, err
		}
		ds.header = data[0]
		ds.data = data[1:]
	} else {
		haveHeader := false
		s := bufio.NewScanner(f)
		for s.Scan() {
			row := strings.Split(s.Text(), "\t")
			if !haveHeader {
				ds.header = row
				haveHeader = true
				continue
			}
			ds.data = append(ds.data, row)
		}
	}

	for i, h := range ds.header {
		if h == "" {
			ds.skiplist[i] = "(empty)"
			continue
		}
		ds.rel.Attrs.Add(funcdep.Attr(h))
	}

	return ds, nil
}

// Analyze a dataset to determine functional dependencies.
func (ds *DataSet) Analyze() {
	// TODO: check more than just pairs of columns

	// for every (ordered) pair of columns,
	// examine dependency between values
	for i := range ds.header {
		if _, skip := ds.skiplist[i]; skip {
			continue
		}
		for j := range ds.header {
			if _, skip := ds.skiplist[j]; skip {
				continue
			}
			if j > i {
				ds.CheckColumnPair(i, j)
			}
		}
	}

	// all of the resulting functional dependencies are
	// simple pairs A->B, A->C, etc. we want all the
	// attributes on the "right" to be combined for
	// each "left" attribute. A->BC etc
	baseFDs := ds.rel.FuncDeps
	ds.rel.FuncDeps = nil

	newFDs := make(map[string]*funcdep.FuncDep)
	for _, fd := range baseFDs {
		key := fd.Left.String()
		if xfd, ok := newFDs[key]; ok {
			xfd.Right.AddAll(fd.Right)
		} else {
			newFDs[key] = fd
		}
	}
	for _, fd := range newFDs {
		ds.rel.FuncDeps = append(ds.rel.FuncDeps, fd)
	}
}

// CheckColumnPair counts data co-occurance for the two columns given.
// If either column (or both) functionally determines the other, then
// the relationship is recorded.
func (ds *DataSet) CheckColumnPair(i, j int) {
	// value_j => set of value_i
	deps := make(map[string]map[string]struct{})

	// value_i => set of value_j
	revdeps := make(map[string]map[string]struct{})

	// read through the data set and track values
	// observed for each pair
	for _, row := range ds.data {
		vi := row[i]
		vj := row[j]

		if _, ok := deps[vj]; !ok {
			deps[vj] = map[string]struct{}{vi: struct{}{}}
		} else {
			deps[vj][vi] = struct{}{}
		}

		if _, ok := revdeps[vi]; !ok {
			revdeps[vi] = map[string]struct{}{vj: struct{}{}}
		} else {
			revdeps[vi][vj] = struct{}{}
		}
	}

	// if all vi are unique to each vj, then j -> i
	uniqueValues := true
	for _, vi := range deps {
		if len(vi) > 1 {
			uniqueValues = false
			break
		}
	}
	if uniqueValues {
		fd := &funcdep.FuncDep{}
		fd.Left.Add(funcdep.Attr(ds.header[j]))
		fd.Right.Add(funcdep.Attr(ds.header[i]))
		ds.rel.FuncDeps = append(ds.rel.FuncDeps, fd)
	}

	///////

	// if all vj are unique to each vi, then i -> j
	uniqueValues = true
	for _, vj := range revdeps {
		if len(vj) > 1 {
			uniqueValues = false
			break
		}
	}

	if uniqueValues {
		fd := &funcdep.FuncDep{}
		fd.Left.Add(funcdep.Attr(ds.header[i]))
		fd.Right.Add(funcdep.Attr(ds.header[j]))
		ds.rel.FuncDeps = append(ds.rel.FuncDeps, fd)
	}

	return
}

func main() {
	excludeList := flag.String("x", "", "comma-separated list of `attributes` to exclude")
	flag.Parse()

	ds, err := ReadData(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	if *excludeList != "" {
		parts := strings.Split(*excludeList, ",")
		for j, p := range parts {
			parts[j] = strings.TrimSpace(p)
		}

		for i, h := range ds.header {
			for _, p := range parts {
				if h == p {
					ds.skiplist[i] = p
					ds.rel.Attrs.Remove(funcdep.Attr(p))
					break
				}
			}
		}
	}
	ds.Analyze()

	fmt.Println(ds.rel.String())

	fmt.Println("---")
	fmt.Println("Candidate Keys:")
	cks := ds.rel.CandidateKeys()
	if len(cks) == 0 {
		cks = ds.rel.CandidateKeysAlt()
	}
	if len(cks) == 0 {
		fmt.Println("No straightforward Candidate Keys -- Need a brute-force search!")
	}
	best := len(ds.rel.Attrs)
	for _, ck := range cks {
		if len(ck) < best {
			best = len(ck)
		}
		fmt.Println("   ", ck)
	}

	if best > 2 {
		fmt.Println("Candidate Keys (Brute-Force):")
		cks = ds.rel.CandidateKeysBF()
		for _, ck := range cks {
			fmt.Println("   ", ck)
		}
	}
}
