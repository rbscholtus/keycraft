package main

import (
	"fmt"

	corpus "github.com/rbscholtus/kb/internal/corpus"
	layout "github.com/rbscholtus/kb/internal/layout"
	"github.com/yassinebenaid/godump"
)

func main() {
	f, err := ParseFlags()
	if err != nil {
		fmt.Println(err)
		return
	}

	godump.Dump(f)

	// Load the corpus from the specified file
	corp, err := corpus.NewFromFile(f.Corpus, "data/corpus/"+f.Corpus)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(corp)

	// Load the layout from the specified file
	lay, err := layout.LoadFromFile("data/layouts/" + f.Layout)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(lay)

	lay.SaveToFile("qwerty.kb")
}
