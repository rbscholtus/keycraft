package main

import (
	"flag"
	"fmt"
)

type Flags struct {
	Layout    string
	Corpus    string
	ShowUsage bool
}

func ParseFlags() (*Flags, error) {
	layout := flag.String("l", "", "layout file to load (located in data/layouts/)")
	// flag.StringVar(layout, "layout", "", "layout file to load (located in data/layouts/)")

	corpus := flag.String("c", "", "corpus file to load (located in data/corpus/)")
	// flag.StringVar(corpus, "corpus", "", "corpus file to load (located in data/corpus/)")

	h := flag.Bool("h", false, "show usage")
	// flag.BoolVar(h, "help", false, "show usage")

	flag.Parse()

	if *h {
		flag.Usage()
		return nil, fmt.Errorf("please specify flags")
	}

	if *layout == "" || *corpus == "" {
		return nil, fmt.Errorf("please specify both a layout file using the -l or --layout flag and a corpus file using the -c or --corpus flag")
	}

	return &Flags{
		Layout: *layout,
		Corpus: *corpus,
	}, nil
}
