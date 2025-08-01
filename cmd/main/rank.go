package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rbscholtus/kb/internal/layout"
	"github.com/urfave/cli/v2"
)

var rankCommand = &cli.Command{
	Name:      "rank",
	Usage:     "Rank multiple layout files with a corpus file",
	ArgsUsage: "<layout files...>",
	Action:    rankAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "corpus",
			Aliases:  []string{"c"},
			Usage:    "specify the corpus file",
			Required: true,
		},
	},
}

func rankAction(c *cli.Context) error {
	layoutFiles := c.Args().Slice()
	corpusFile := c.String("corpus")

	if len(layoutFiles) < 1 {
		return fmt.Errorf("at least one layout file is required")
	}

	fmt.Printf("Ranking layouts: %v with corpus: %s\n", layoutFiles, corpusFile)
	doRankings(corpusFile)
	return nil
}

type LayoutScore struct {
	Name  string
	Score float64
}

func doRankings(corpusFile string) {
	// Read corpus
	corpus, err := layout.NewCorpusFromFile("corpus", filepath.Join("data", "corpus", corpusFile))
	if err != nil {
		fmt.Println(err)
		return
	}

	// Read layouts
	layoutFiles, err := os.ReadDir("data/layouts")
	if err != nil {
		fmt.Println(err)
		return
	}

	var layouts []*layout.SplitLayout
	for _, file := range layoutFiles {
		if !strings.HasSuffix(file.Name(), ".klf") {
			continue
		}
		layout, err := layout.NewLayoutFromFile(file.Name(), filepath.Join("data", "layouts", file.Name()))
		if err != nil {
			fmt.Println(err)
			continue
		}
		layouts = append(layouts, layout)
	}

	// Analyze layouts
	var metrics = make(map[string][]float64)
	for _, lay := range layouts {
		analyser := layout.NewAnalyser(lay, corpus)
		for metric, value := range analyser.Metrics {
			metrics[metric] = append(metrics[metric], value)
		}
	}

	// Calculate median and interquartile range for each metric
	medians := make(map[string]float64)
	iqr := make(map[string]float64)
	for metric, values := range metrics {
		sort.Float64s(values)
		medians[metric] = layout.Median(values)
		q1, q3 := layout.Quartiles(values)
		iqr[metric] = q3 - q1
	}

	// Scale and sum metrics
	var layoutScores []LayoutScore
	for _, lay := range layouts {
		analyser := layout.NewAnalyser(lay, corpus)
		score := 0.0
		for metric, value := range analyser.Metrics {
			scaledValue := (value - medians[metric]) / iqr[metric]
			// You can choose to penalize or reward certain metrics by multiplying with a factor
			score += scaledValue
		}
		layoutScores = append(layoutScores, LayoutScore{lay.Name, score})
	}

	// Rank layouts
	sort.Slice(layoutScores, func(i, j int) bool {
		return layoutScores[i].Score < layoutScores[j].Score
	})

	// Print results
	fmt.Println("Ranking:")
	for i, layoutScore := range layoutScores {
		fmt.Printf("%d. %s (Score: %.2f)\n", i+1, layoutScore.Name, layoutScore.Score)
	}
}
