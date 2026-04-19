# Architecture

High-level map of the keycraft domain model and data flow. For the `generate` subcommand specifically, see [`GENERATION.md`](./GENERATION.md). For user-facing features and metrics, see [`../README.md`](../README.md).

## Data flow

```
┌──────────────────┐        ┌──────────────────┐
│ data/corpus/*.json│───────▶│ Corpus            │
└──────────────────┘        │ (corpus.go)       │
                            └────────┬─────────┘
┌──────────────────┐                 │
│ data/layouts/*.klf│──┐              ▼
└──────────────────┘  │    ┌──────────────────┐
                      ├──▶│ Analyser          │───▶ per-layout metrics
┌──────────────────┐  │    │ (analyser.go,     │
│ data/config/*.gen │──┘   │  analyse.go)      │
└──────────────────┘       └─────────┬────────┘
                                     │
                       ┌─────────────┼──────────────┬───────────────┐
                       ▼             ▼              ▼               ▼
                  ┌────────┐   ┌──────────┐   ┌──────────┐    ┌──────────┐
                  │ View   │   │ Ranking  │   │ Optimiser│    │ Generator│
                  │view.go │   │ranking.go│   │ (BLS)    │    │generator │
                  └───┬────┘   └────┬─────┘   │optimize, │    │  .go     │
                      │             │         │ bls*.go  │    └────┬─────┘
                      │             │         └────┬─────┘         │
                      │             │              ▼               ▼
                      │             │         ┌──────────┐    ┌──────────┐
                      │             │         │ Scorer   │◀──▶│ Weights  │
                      │             │         │scorer.go │    │weights.go│
                      │             │         └──────────┘    └──────────┘
                      ▼             ▼                              │
                  ┌─────────────────────────────────────────────────┘
                  │
                  ▼
          ┌──────────────────┐
          │ internal/tui/*.go │ ──▶ stdout / HTML / CSV
          └──────────────────┘
```

## Packages

### `cmd/keycraft`

Thin CLI glue — one file per subcommand (`corpus.go`, `view.go`, `analyse.go`, `rank.go`, `flip.go`, `optimize.go`, `generate.go`). `main.go` assembles the root command; `flags.go` centralises flag declarations; `helpers.go` hosts shared load/parse routines. No domain logic lives here.

### `internal/keycraft`

The domain. Stable types:

| Type            | File                          | Role                                                                         |
| --------------- | ----------------------------- | ---------------------------------------------------------------------------- |
| `Layout`        | `layout.go`, `keys.go`        | Physical/logical keyboard layout; swap/pin primitives.                      |
| `Corpus`        | `corpus.go`                   | JSON-loaded n-gram frequencies (unigrams, bigrams, trigrams, words).        |
| `Analyser`      | `analyser.go`, `analyse.go`   | Runs metrics (SFB, LSB, FSB, HSB, trigram categories, loads) over a layout. |
| `Scorer`        | `scorer.go`                   | Aggregates metrics with user-supplied weights into a single score.          |
| `Weights`       | `weights.go`                  | Weight vector; default + loaded from file.                                   |
| `Ranking`       | `ranking.go`                  | Compute + sort scores across many layouts; compute medians/IQR for norm.    |
| `Optimiser`     | `optimize.go`, `bls*.go`      | Breakout Local Search over layout key swaps.                                |
| `Generator`     | `generator.go`                | Enumerate permutations of free positions from a `.gen` config.              |
| `Targets`       | `targets.go`                  | Target finger/row load profiles used by scorer penalties.                    |
| `Random`        | `random.go`                   | Single seeded RNG source for all randomness. Reproducibility lives here.    |
| `View`          | `view.go`                     | Single-layout inspection data (feeds `internal/tui/view.go`).               |

### `internal/tui`

Presentation layer. Uses [`jedib0t/go-pretty/v6`](https://github.com/jedib0t/go-pretty) for tables. One file per command (`view.go`, `analyse.go`, `ranking.go`, `corpus.go`, `optimize.go`) plus shared helpers (`format.go`, `styles.go`). Keep rendering concerns here; do not couple `internal/keycraft` to table formatting.

### `data/`

- `corpus/*.json` — pre-computed n-gram data. Only re-generate when the source corpus changes.
- `layouts/*.klf` — layout definitions.
- `layouts2/` — secondary layout collection.
- `config/*.gen` — generator configs (see `docs/GENERATION.md`).
- `gen_layouts/` — generator output; gitignored.

## Cross-cutting concerns

### Determinism / RNG

All randomness flows through `internal/keycraft/random.go`. The optimiser (BLS) and generator both seed from this source so that runs with the same `--seed` reproduce. **Do not** use `math/rand` or `crypto/rand` directly elsewhere in the module — it breaks reproducibility.

### Error handling

Domain errors are wrapped with context (`fmt.Errorf("…: %w", err)`) at package boundaries so CLI output shows a readable trace. Panics are reserved for programmer errors (e.g., impossible enum values); the `Distance` lookup is an example of a historical panic that was converted to a wrapped error.

### Concurrency

The optimiser supports parallel runs — multiple BLS instances share a precomputed reference-stats snapshot rather than recomputing per goroutine. See `internal/keycraft/bls_parallel_test.go` for the contract. Generator uses the same parallelism primitive.

## Extension points

### Adding a metric

1. Compute it in `analyser.go` (write to the `Analyse` result struct).
2. Add a weight entry in `weights.go` (default value) and wire into `scorer.go`.
3. Expose it in the relevant TUI file (`internal/tui/view.go`, `analyse.go`, `ranking.go`).
4. Document in `README.md` under *Supported Metrics*.

### Adding a layout format

Currently only `.klf` is supported (`layout.go:ParseLayoutFile` or equivalent). To add another format, implement parsing that yields the same `Layout` struct and branch on file extension in the CLI loader.

### Adding a corpus

Produce a JSON file matching the schema used by existing files in `data/corpus/`. The corpus builder is not currently checked in — use an external script. The default corpus is set in `cmd/keycraft/flags.go`.

## Build + release

Two workflows in `.github/workflows/`:

### `go.yml` — release pipeline

Triggers on pushed `v*` tags (and PRs to `main` for build/test only). On a tag push, CI builds Linux/macOS/Windows binaries + `data.tar.gz` and uploads them to a GitHub Release via `softprops/action-gh-release`. The release body is *not* auto-generated — it is whatever was set when the release was created (see `CONTRIBUTING.md` §Releasing for the `gh release create --notes-file` flow that sources notes from `CHANGELOG.md`).

### `static.yml` — docs deployment

Triggers on push to `main`. Builds the binary and regenerates `docs/index.html` by sandwiching the ranking output:

```sh
(cat docs/header.html; ./keycraft r --metrics all --output html; cat docs/footer.html) > docs/index.html
```

If the regenerated file differs from what's in the tree, CI commits it back to `main` with `[skip ci]` (the marker prevents an infinite loop, since the bot's own commit would otherwise re-trigger the workflow). Then it uploads the entire `docs/` folder as a Pages artifact and deploys to https://rbscholtus.github.io/keycraft/ via `actions/deploy-pages@v4`.

Pages is configured via the workflow build type (`build_type: workflow`), not the legacy "deploy from branch" setting. The HTML wrapper styles (light/dark theme, sticky-sortable headers) live in `docs/header.html` + `docs/footer.html`; the body table is whatever `keycraft r --output html` emits — keep its `<table class="keycraft-ranking-table">` selector intact, since the JS in `footer.html` queries on it.
