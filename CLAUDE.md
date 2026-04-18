# CLAUDE.md

Orientation for AI coding sessions working on `keycraft`. See `README.md` for user-facing documentation.

## What this is

CLI tool (single binary) for analysing, ranking, generating, and optimising keyboard layouts. Written in Go. The user runs it against a text corpus plus one or more layout files and gets tables/HTML/CSV output describing typing metrics (SFB, LSB, finger/row load, trigram patterns, etc.).

## Commands

```sh
go build -o keycraft ./cmd/keycraft   # build
go test ./...                         # tests
golangci-lint run                     # lint (not enforced in CI)
go mod tidy                           # after dep changes
```

Binary subcommands (see `cmd/keycraft/main.go`): `corpus`, `view`, `analyse`, `rank`, `flip`, `optimize`, `generate`.

## Layout

- `cmd/keycraft/` — one file per CLI subcommand, plus `main.go`, `flags.go`, `helpers.go`. Command wiring is thin; domain logic lives in `internal/`.
- `internal/keycraft/` — domain. Key types:
  - `Layout` (`layout.go`, `keys.go`) — keyboard layout model.
  - `Corpus` (`corpus.go`) — text frequency data, JSON-persisted.
  - `Analyser` (`analyser.go`, `analyse.go`) — metric computation.
  - `Scorer` + `Weights` (`scorer.go`, `weights.go`) — weighted metric aggregation for ranking/optimising.
  - `Ranking` (`ranking.go`) — compare many layouts.
  - Optimiser (`optimize.go`, `bls.go`, `bls_utils.go`, `bls_logger.go`) — **Breakout Local Search (BLS)** algorithm.
  - Generator (`generator.go`) — synthesise new layouts from a `.gen` config.
  - `Random` (`random.go`) — seeded RNG. All randomness flows through here so runs are reproducible.
  - `Targets` (`targets.go`) — target finger/row load profiles.
  - `View` (`view.go`) — single-layout inspection.
- `internal/tui/` — table/HTML/CSV rendering per command (`go-pretty/v6`). Keep presentation concerns here, not in `internal/keycraft/`.
- `data/`
  - `corpus/*.json` — pre-computed n-gram frequencies per corpus.
  - `layouts/*.klf` — layout definitions (`.klf` = keycraft layout file).
  - `config/*.gen` — generator configs.
  - `gen_layouts/` — output of generator runs (gitignored).
  - `layouts2/` — secondary layout collection.

## Pipeline mental model

```
corpus JSON ─▶ Analyser ─▶ per-layout metrics ─┬─▶ Scorer+Weights ─▶ Ranking/Optimiser/Generator ─▶ TUI
                                               └─▶ View (single layout)
```

## Conventions worth knowing

- Public types use doc comments; keep them. Most `internal/keycraft` types are exported within the module — that's intentional for test visibility.
- Tests sit next to sources (`*_test.go`). Use `-bench` suffixed files for benchmarks.
- CLI uses `urfave/cli/v3`. Flags are declared in `cmd/keycraft/flags.go` and composed per command.
- Tables come from `jedib0t/go-pretty/v6`. Output styles live in `internal/tui/styles.go` and `format.go`.
- Version string is a const in `cmd/keycraft/main.go` (`Version:` field on the root `cli.Command`). Bump it when cutting a release.
- CI triggers on tags `v*` — pushing a tag auto-builds cross-platform binaries and creates the GitHub Release. See `.github/workflows/go.yml`.

## Roadmap

`todo.txt` is the live roadmap — short checklist, edit freely. Items are grouped by target version.

## When editing

- Don't duplicate metric formulas; extend `analyser.go`/`scorer.go` instead.
- RNG must go through `internal/keycraft/random.go`. Do not call `math/rand` or `crypto/rand` directly in domain code.
- `.bak` files in the tree are scratch from earlier refactors; ignore them (don't compile, don't edit).
- Docs regeneration (`docs/index.html`) is a manual step; CI commits show `[skip ci]` on those — leave that convention alone.
