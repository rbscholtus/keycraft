# Contributing

Thanks for your interest in keycraft. This is a personal project; contributions are welcome.

## Prerequisites

- **Go 1.26+** (see `go.mod`)
- **golangci-lint** (optional but recommended; version 2.x config in `.golangci.yml`)
- Git

## Build, test, lint

```sh
go build -o keycraft ./cmd/keycraft
go test ./...
golangci-lint run
```

The binary expects `data/` to be next to it (corpus, layouts, configs). Run from the repo root or a directory that contains `data/`.

## Branch and commit conventions

- Branch names: `feature/<short-name>`, `fix/<short-name>`, `refactor/<short-name>`.
- Commit messages: short imperative first line ("Add X", "Fix Y", "Refactor Z"). Multi-line bodies welcome when explaining non-obvious decisions.
- Auto-regenerated `docs/index.html` commits use `[skip ci]` so they don't trigger the build.

## Pull requests

- Keep PRs focused — one feature or fix per PR when possible.
- Run tests and lint locally before pushing.
- CI (`.github/workflows/go.yml`) builds and tests on push to `main` / PRs to `main`. It does not run lint.
- Pushing a `v*` tag triggers a cross-platform release build. Do this only from `main`.

## Adding content

### A new layout

1. Drop a `.klf` file into `data/layouts/`.
2. Confirm it loads: `./keycraft view <name>`.
3. It is picked up automatically by `rank`/`optimize` — no code change needed. Naming affects how it is used though: filenames that start with `_` or contain `-flipped`, `-best`, or `-opt` are treated as non-reference layouts. They are still loaded and ranked, but excluded from the median/IQR normalization baseline used for scoring (see `isReferenceLayout()` in `internal/keycraft/scorer.go`). Use a plain name (e.g. `colemak.klf`) for a layout that should contribute to the baseline.
4. The next push to `main` also auto-generates `docs/layouts/<name>.html` for it (see `docs/ARCHITECTURE.md` §`static.yml`); the ranking page links there.

### A new corpus

1. Drop a plain-text `.txt` file into `data/corpus/`. On the first run that references it, `NewCorpusFromFile` (`internal/keycraft/corpus.go`) parses the text and writes a `<name>.txt.json` cache next to it; subsequent runs load from the cache. Pass `--coverage` to truncate by cumulative word frequency.
2. To make it the default, name the source file `default.txt` — the `--corpus` flag defaults to `default.txt`, so renaming overrides the existing default without editing code.

### A new metric

1. Extend `internal/keycraft/analyser.go` to compute it.
2. Wire weights in `internal/keycraft/scorer.go` and `weights.go`.
3. Add a column in `internal/tui/ranking.go` / relevant TUI file.
4. Update `README.md` metric documentation.

## Architecture

See `docs/ARCHITECTURE.md` for the domain model and data flow. See `CLAUDE.md` for a quick orientation geared at AI coding sessions (also useful for humans).

## Releasing (maintainer)

1. Update `CHANGELOG.md` — move items from `[Unreleased]` into a new versioned section.
2. Bump the `Version` constant in `cmd/keycraft/main.go`.
3. Commit.
4. Tag: `git tag -a vX.Y.Z -m "vX.Y.Z"` and `git push origin vX.Y.Z`.
5. CI publishes the release automatically (binaries + `data.tar.gz`).
