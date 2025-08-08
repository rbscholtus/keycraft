package main

import (
	"fmt"
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

	corp, err := loadCorpus(c)
	if err != nil {
		return err
	}

	weights, err := parseWeights(c.String("weights"))
	if err != nil {
		return err
	}

	style := c.String("style")

	return ly.DoRankings(corp, layoutDir, weights, style)
}

// parseWeights parses a string of weighted metrics into a map of metric names to weights.
// The input string is expected to be in the format "metric1=value1,metric2=value2,...".
// If a metric is not specified in the input string, its weight will default to the value specified in the weights map.
// Currently, ALT and ROL have default weights of -1, because they are "positive penalties".
func parseWeights(weightsStr string) (map[string]float64, error) {
	weights := map[string]float64{
		"ALT": -1,
		"ROL": -1,
		"ONE": -1,
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
