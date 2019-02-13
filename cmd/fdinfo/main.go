// Command fdinfo reads in functional dependencies and lists various properties that can be inferred.
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/joiningdata/funcdep"
)

func main() {
	nosep := flag.Bool("n", false, "use single-character attribute names (no separator)")
	delim := flag.String("d", ",", "use `separator` between attribute names")
	flag.Parse()

	if *delim != "" {
		funcdep.AttrSep = *delim
	}
	if *nosep {
		funcdep.AttrSep = ""
	}

	var r io.ReadCloser = os.Stdin
	if fn := flag.Arg(0); fn != "" {
		f, err := os.Open(fn)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		r = f
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	r.Close()

	rel, err := funcdep.RelationFromString(string(data))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	fmt.Println(rel)

	fmt.Println("Candidate Keys:")
	cks := rel.CandidateKeys()
	if len(cks) == 0 {
		cks = rel.CandidateKeysAlt()
	}
	if len(cks) == 0 {
		fmt.Println("No straightforward Candidate Keys -- Need a brute-force search!")
	}
	for _, ck := range cks {
		fmt.Println("   ", ck)
	}
}
