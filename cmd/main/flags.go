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
	AcceptWorse string
	Generations int
}

func ParseFlags() (*Flags, error) {
	help := flag.Bool("h", false, "show usage")
	layout := flag.String("l", "", "layout file to load (located in data/layouts/)")
	corpus := flag.String("c", "", "corpus file to load (located in data/corpus/)")
	acceptWorse := flag.String("accept-worse", "temp", "accept worse function: always, drop-slow, temp, drop-fast, cold, or never")
	generations := flag.Int("gens", 99, "number of generations (must be above 0)")

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
	if !contains(validAcceptWorse, *acceptWorse) {
		return nil, fmt.Errorf("invalid accept worse function: %s. Must be one of: %v", *acceptWorse, validAcceptWorse)
	}

	// Validate number of generations
	if *generations <= 0 {
		return nil, fmt.Errorf("number of generations must be above 0. Got: %d", *generations)
	}

	return &Flags{
		Layout:      *layout,
		Corpus:      *corpus,
		AcceptWorse: *acceptWorse,
		Generations: *generations,
	}, nil
}

func contains(s []string, e string) bool {
	return slices.Contains(s, e)
}
