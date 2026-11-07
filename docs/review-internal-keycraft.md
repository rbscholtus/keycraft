# Review: `internal/keycraft` (2026-11-07)

Manual review of the whole package (~6k lines of non-test code), focused on the
optimiser, scorer, analyser and layout. `go vet` clean, tests pass at time of review.
Nothing has been fixed yet.

## Bugs

1. **Pins file can crash `optimize`** — `bls_utils.go:239-244`.
   Without `--pins-file`, empty `~` keys are pinned by default. With a pins file, its
   pins replace that default, so an empty key marked `.` becomes swappable and the
   first swap panics (`can't swap unused key`, `layout.go:194`).
   *Fix:* always pin `r == 0` positions after loading the file. Add a regression test.

2. **`generate --seed 0` is only random for the first layout** — `generator.go:732`.
   The RNG seed is `seed+permIndex`; only a combined 0 falls back to the clock. With
   seed 0, layouts 1..N get fixed seeds 1..N every run. Also `--seed s` layout `i`
   reuses the shuffle of `--seed s+1` layout `i-1`.
   *Fix:* resolve 0 to a timestamp once, then derive per-layout seeds (hash, or draw
   from one RNG). Add a regression test.

## Correctness / consistency

3. **Optimiser bypasses the shared RNG** — `bls.go:5, 173` use `math/rand` (v1)
   directly; `CLAUDE.md` requires going through `random.go`.
4. **`--generations` miscounts** — `bls.go:237, 360, 523`. `MaxIterations` is
   documented as local optima visited, but `state.iteration` also increments per
   descent swap and per perturbation move, so runs end early. The
   `iteration % ReportInterval` progress check fires erratically for the same reason.
5. **Optimiser and ranking score differently** — `scorer.go`.
   - Optimiser keeps only trigrams covering 98% of occurrences (line 174) but divides
     by the full `TotalTrigramsCount` and compares against medians from all trigrams,
     skewing ALT, RED, roll metrics, FLW and IN:OUT ~2% low.
   - Optimiser drops weights with |w| <= 0.01 (line 73); `computeScores` keeps any
     non-zero weight (line 457).
6. **Normalisation panics with one reference layout** — `scorer.go:490`. `Quartiles`
   calls `Median` on an empty slice (index `[-1]`).
7. **IN:OUT can be infinite** — `analyser.go:512` divides by zero when there are no
   outward rolls; Inf/NaN then flows into medians and scores.
8. **Possible data race on partly filled `TargetLoads`** — `NewAnalyser` and
   `OptimizeLayoutBLS` fill nil fields in place on the caller's struct, and
   `LoadAnalysers` calls `NewAnalyser` from several goroutines with one pointer.
   Current callers appear to pass fully filled structs.
9. **Copy-paste slip** — `analyser.go:737, 887` check `ma.Custom` instead of
   `ma2.Custom`. Harmless today; would panic if a bigram were in both lists.

## Cleanup

- `identifyProblematicKeys` (`bls.go:643`) is test-only; the live code uses
  `identifyProblematicKeys2`. Delete the old one and rename.
- `NewScorer` duplicates the filtering loop in `ComputeReferenceStats`; have it call
  `ComputeReferenceStats` then `NewScorerWithStats`.
- `AnalyseLayouts` and `ViewLayouts` are nearly identical.
- The score cache never evicts, so memory grows for the whole optimiser run. Consider
  a bounded cache.
