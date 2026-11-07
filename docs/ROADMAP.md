# Keycraft Roadmap

## v0.7

- Some feature(s) from the backlog

### Fixes

- Clean up `interface with *Layout and bool` — currently ugly
- Table rendering issues
- Number of rows split by bigram / trigram counts
- `optimize --pins-file` panics when an empty `~` key is left unpinned
  (see `docs/review-internal-keycraft.md` #1)
- `generate --seed 0` is only random for the first layout; per-layout seeds overlap
  across runs (see `docs/review-internal-keycraft.md` #2)

### CI / Release

- **Auto-populate release notes**: Configure `softprops/action-gh-release` in `go.yml`
  to include a release body (e.g. from a `RELEASE_NOTES.md` template or tag annotation)
  so GitHub Releases are not created empty.

---

## Experimental: API / MCP / Desktop (branch: `feat/api-layer`)

A transport-agnostic `internal/api.Service` layer is being prototyped as the
foundation for non-CLI frontends. Three adapters are under exploration:

- **`keycraft-serve`** — localhost HTTP API (intended for web UI development via
  a Vite proxy, or any same-machine client)
- **`keycraft-mcp`** — MCP stdio server so LLM clients (Claude Desktop, Cursor,
  etc.) can drive keycraft directly as a tool
- **`keycraft-wails`** — desktop GUI using [Wails](https://wails.io), with the
  frontend calling Go methods via generated TypeScript bindings

All three share the same `api.Service` so business logic is not duplicated across
transports. The branch is experimental and not reviewed for production use.

---

## Backlog

### Metrics

- **ICC — Index Centre Column** (#11): Combined centre-column usage (C5+C6) as a new
  metric. Requested as a lightweight alternative to per-column breakdown.

- **Weighted finger bigrams**: New variant of FSB (and possibly SFB/LSB) that weights
  bigram counts by finger strength — so the same bigram costs more if it falls on a
  weaker finger (ring, pinky). Configurable penalty multipliers per finger.

- **Unified finger weights overlay**: Generalise the current `POH` (Pinky Off Home)
  penalty into a single per-finger, per-position weight matrix covering all fingers.
  ICC and POH become outputs of this one config rather than separate mechanisms.
  Relates to weighted finger bigrams above.

### Analysis

- **Bigram table in `analyse`** (#13): A combined bigram table analogous to the existing
  trigram table — with a metric-class column, colour highlights for bad patterns, and a
  configurable row limit.

- **`compare` command**: Side-by-side delta view between two specific layouts.

- **Analysis by trigram decomposition**: Carpalx-style effort model driven by trigrams.

### Optimiser

- Progress bar for optimisation
- Remove `--max-time` flag
- Improve seed generation; verify seed reproducibility
- Improve optimiser hints output
- Expose optimiser strategy params
- `--max-seed` as a global hidden flag
- Progress bar / slice optimiser progress

---

## Later / Ideas

- Alt fingering support
- Magic keys
- Multiple layers
- Analysis in Carpalx style
