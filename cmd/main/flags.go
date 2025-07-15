package main

import (
	"flag"
	"fmt"
	"slices"
)

type Flags struct {
	ShowUsage   bool
	Layout      string
	Corpus      string
	Optimize    bool
	Pins        string
	Generations uint
	AcceptWorse string
}

func ParseFlags() (*Flags, error) {
	help := flag.Bool("h", false, "show usage")
	layout := flag.String("l", "", "layout file to load (located in data/layouts/)")
	corpus := flag.String("c", "", "corpus file to load (located in data/corpus/)")
	optimize := flag.Bool("o", false, "optimize the layout (default false).")
	pins := flag.String("p", "", "file containing keys the optimiser cannot move (located in data/pins/)")
	generations := flag.Uint("g", 99, "number of generations (must be above 0)")
	acceptWorse := flag.String("f", "temp", "accept worse function: always, drop-slow, temp, drop-fast, cold, or never")

	flag.Parse()

	if *help {
		flag.Usage()
		return nil, fmt.Errorf("please specify flags")
	}

	if *layout == "" || *corpus == "" {
		return nil, fmt.Errorf("please specify both a layout file using the -l flag and a corpus file using the -c flag")
	}

	// Validate accept worse function
	validAcceptWorse := []string{"always", "drop-slow", "temp", "drop-fast", "cold", "never"}
	if !slices.Contains(validAcceptWorse, *acceptWorse) {
		return nil, fmt.Errorf("invalid accept worse function: %s. Must be one of: %v", *acceptWorse, validAcceptWorse)
	}

	// Validate number of generations
	if *generations <= 0 {
		return nil, fmt.Errorf("number of generations must be above 0. Got: %d", *generations)
	}

	return &Flags{
		Layout:      *layout,
		Corpus:      *corpus,
		Optimize:    *optimize,
		Pins:        *pins,
		Generations: *generations,
		AcceptWorse: *acceptWorse,
	}, nil
}
