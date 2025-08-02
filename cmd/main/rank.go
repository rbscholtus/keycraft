package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	ly "github.com/rbscholtus/kb/internal/layout"
	"github.com/urfave/cli/v2"
)

// the "rank" cli command for ranking all layouts
var rankCommand = &cli.Command{
	Name:   "rank",
	Usage:  "Rank layout files in data/layouts with a corpus file",
	Action: rankAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "corpus",
			Aliases:  []string{"c"},
			Usage:    "specify the corpus file",
			Required: true,
		},
		&cli.StringFlag{
			Name:    "weights",
			Aliases: []string{"w"},
			Usage:   "specify weights for metrics (e.g. sfb=3.0,lsb=2.0)",
			Value:   "",
		},
	},
}

// rankAction is the action function for the rank command.
func rankAction(c *cli.Context) error {
	// Check if any arguments were specified (none are expected)
	if c.Args().Present() {
		return fmt.Errorf("invalid argument(s) specified: %v", c.Args().Slice())
	}
	corpusPath := filepath.Join("data", "corpus", c.String("corpus"))

	weights, err := parseWeights(c.String("weights"))
	if err != nil {
		return err
	}

	fmt.Printf("Ranking layouts in data/layouts with %s and weights: %v\n", corpusPath, weights)
	return doRankings(corpusPath, "data/layouts", weights)
}

// LayoutScore represents a layout's score, including its name and penalty value.
type LayoutScore struct {
	Name    string
	Penalty float64
}

// doRankings performs the ranking of layouts based on the corpus and weights.
func doRankings(corpusPath, layoutsDir string, weights map[string]float64) error {
	// Read corpus
	corpus, err := ly.NewCorpusFromFile("corpus", corpusPath)
	if err != nil {
		return fmt.Errorf("error finding corpus %v: %v", corpusPath, err)
	}

	// Read layouts
	layoutFiles, err := os.ReadDir(layoutsDir)
	if err != nil {
		return fmt.Errorf("error finding layout files in %v: %v", layoutsDir, err)
	}

	// Analyze layouts
	var analysers = make([]*ly.Analyser, 0)
	var metrics = make(map[string][]float64)
	for _, file := range layoutFiles {
		if !strings.HasSuffix(file.Name(), ".klf") {
			continue
		}
		layoutPath := filepath.Join("data", "layouts", file.Name())
		layout, err := ly.NewLayoutFromFile(file.Name(), layoutPath)
		if err != nil {
			fmt.Println(err)
			continue
		}
		analyser := ly.NewAnalyser(layout, corpus)
		analysers = append(analysers, analyser)
		for metric, value := range analyser.Metrics {
			metrics[metric] = append(metrics[metric], float64(value))
		}
	}

	// Calculate median and interquartile range for each metric
	// This is used to normalize/scale the metric values
	medians := make(map[string]float64)
	iqr := make(map[string]float64)
	for metric, values := range metrics {
		sort.Float64s(values)
		medians[metric] = ly.Median(values)
		q1, q3 := ly.Quartiles(values)
		iqr[metric] = q3 - q1

	}

	// Scale and sum metrics
	// Calculate the penalty for each layout based on the scaled metric values and weights
	var layoutPenalties []LayoutScore
	for _, analyser := range analysers {
		penalty := 0.0
		for metric, value := range analyser.Metrics {
			weight, ok := weights[metric]
			if !ok {
				// default weight, if unspecified
				weight = 1.0
			}
			var scaledValue float64
			if iqr[metric] == 0 {
				scaledValue = 0
			} else {
				scaledValue = (value - medians[metric]) / iqr[metric]
			}
			penalty += weight * scaledValue
		}
		layoutPenalties = append(layoutPenalties, LayoutScore{analyser.Layout.Name, penalty})
	}

	// Rank layouts
	sort.Slice(layoutPenalties, func(i, j int) bool {
		return layoutPenalties[i].Penalty < layoutPenalties[j].Penalty
	})

	// Print results
	fmt.Println("Ranking:")
	for i, layoutPenalty := range layoutPenalties {
		fmt.Printf("%d. %s (Penalty: %.2f)\n", i+1, layoutPenalty.Name, layoutPenalty.Penalty)
	}

	return nil
}

func parseWeights(weightsStr string) (map[string]float64, error) {
	weights := map[string]float64{
		"ALT": -1,
		"ROL": -1,
	}

	if weightsStr == "" {
		return weights, nil
	}

	weightsStr = strings.ToUpper(strings.TrimSpace(weightsStr))
	for pair := range strings.SplitSeq(weightsStr, ",") {
		parts := strings.Split(pair, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid weights format")
		}
		metric := strings.TrimSpace(parts[0])
		weight, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid weight value for metric %s", metric)
		}
		weights[metric] = weight
	}
	return weights, nil
}
