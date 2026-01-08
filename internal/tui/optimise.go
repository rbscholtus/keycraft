package tui

import (
	"fmt"

	kc "github.com/rbscholtus/keycraft/internal/keycraft"
)

// RenderOptimise renders optimization results to stdout.
// Displays analysis comparison and ranking table for original vs optimized layouts.
func RenderOptimise(result *kc.OptimiseResult, rankingResult *kc.RankingResult, weights *kc.Weights) error {
	// First, show analysis comparison using view rendering
	viewResult := &kc.ViewResult{
		Analysers: []*kc.Analyser{},
	}

	// Find analysers for original and best layouts from ranking result
	for _, score := range rankingResult.Scores {
		if score.Name == result.OriginalLayout.Name || score.Name == result.BestLayout.Name {
			viewResult.Analysers = append(viewResult.Analysers, score.Analyser)
		}
	}

	// Render view comparison
	if err := RenderView(viewResult); err != nil {
		return fmt.Errorf("failed to render layout analysis: %v", err)
	}

	// Then render ranking table
	displayOpts := RankingDisplayOptions{
		OutputFormat:   OutputTable,
		MetricsOption:  MetricsWeighted,
		ShowWeights:    true,
		Weights:        weights,
		DeltasOption:   DeltasCustom,
		BaseLayoutName: result.OriginalLayout.Name,
	}

	if err := RenderRankingTable(rankingResult, displayOpts); err != nil {
		return fmt.Errorf("failed to render layout rankings: %v", err)
	}

	return nil
}
