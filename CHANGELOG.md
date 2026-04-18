# Changelog

All notable changes to this project are documented here. Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versioning follows [SemVer](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

Items carried over from `todo.txt` for `v0.6.0` final:
- Max-seed global hidden flag.
- Clean up `*Layout + bool` interface.
- Progress bar / slice-optimiser progress polish.

## [0.6.0-beta.1] - 2026-04-19

First pre-release of the `generate` command and several ranking/analyser improvements.

### Added
- `generate` command: synthesise layouts from a `.gen` config, and apply parallel optimisation. See `data/config/example.gen` for a starting point.
- Progress bar for parallel optimisation.
- Seeded RNG via `internal/keycraft/random.go`; all random behaviour is now reproducible.
- Layouts: 7 from cyanophage, plus `afterburner` and `racket`.
- Corpus: `reddit_small`; apostrophe support in tokenised words.
- Ranking table title now includes the corpus name.
- Wrapped error messages across CLI actions for clearer failure traces.
- Panic on unknown `Distance` lookup replaced with a wrapped error.

### Changed
- `-best` flag renamed to `-opt` on optimise/rank commands.
- Default corpus switched from `default` to `reddit_small`.
- Stats table header `Cum%` → `Cumul%`.
- Trigram output order: Redirects moved before other trigrams; `FLW` and `I:O` swapped.
- Internal rename `optimise.go` → `optimize.go` (both in `internal/keycraft/` and `internal/tui/`).
- Go toolchain: `1.25.5` → `1.26.2`.

### Fixed
- Incorrect "Weak redirect" details.
- Double character sometimes appearing in generated layouts.
- Reference-layout filtering in `computeMediansAndIQR` and scorer normalisation.
- Various typos, linter nits, and test cleanups.

### Infra
- CI action name and data-upload fixes; copyright bumped to 2025–2026.
- Branch name corrected from `feeature/generate2` → `feature/generate2`.
- New docs: `CLAUDE.md`, `CHANGELOG.md`, `CONTRIBUTING.md`, `docs/ARCHITECTURE.md`.

## [0.5.0] - 2026-01-18

Previous release. See `git log v0.4.0..v0.5.0` for details.

## [0.4.0] and earlier

See `git tag --sort=-v:refname` and `git log <tag>` for historical releases.

[Unreleased]: https://github.com/rbscholtus/keycraft/compare/v0.6.0-beta.1...HEAD
[0.6.0-beta.1]: https://github.com/rbscholtus/keycraft/compare/v0.5.0...v0.6.0-beta.1
[0.5.0]: https://github.com/rbscholtus/keycraft/releases/tag/v0.5.0
