# Keycraft

Keycraft is a Golang-based command-line utility for analyzing and optimizing keyboard layouts. It provides layout visualizations, metric analysis (bigram, trigram, skipgram), layout ranking, and optimization tools to help you design your "ideal" layout.

## Quick start

```bash
# Install (requires golang on your system)
go install github.com/rbscholtus/keycraft/cmd/keycraft@latest

# View a built-in layout
keycraft view qwerty.klf

# Analyse multiple layouts
keycraft analyse colemak.klf dvorak.klf

# Rank built-in layouts with custom weights (making SFB the dominant metric)
keycraft rank --weights sfb=-1000

# Optimise the qwerty layout but keeping its signature keys in place
keycraft optimise --pins qwerty qwerty.klf
```

## Features

- Supports 4x6+3 (x2) layouts, and the layout types row-staggered, ortholinear, column-staggered
- Supports metrics that use the Euclidian distance for each physical layout type
- Has 30+ layouts out of the box, based on https://getreuer.info/posts/keyboards/alt-layouts/stats.html
- Supports corpuses in txt format; supports an internal cache for fast loading
- Supports a default corpus, eliminating the need to specify the corpus for every command
- Has corpus files out of the box for MonkeyRacer, Shai, and AKL
- Supports viewing one or more layouts side-by-side with:
  - Graphic of the physical layout (row, ortho, col)
  - Hand and row balance statistics
  - Overview of 16 key metrics
- Supports analysing one or more layouts side-by-side with detailed tables for each metric
- Supports scoring and ranking some or all layouts with:
  - User-defined weights for all metrics, using a weights file
  - User-defined weights on the command-line, useful to easily override weights in the weights file
  - Supports a default weights file, eliminating the need to specify the weights for every command
  - Normalizes metrics using stable scaling (using the median and inter-quartile range of each metric)
  - Deltas between rows for comparing how metrics differ between layouts
  - Deltas between each layout and the median, or a given layout
- Supports optimising a layout with:
  - Pin specific keys using a .pin file and/or specifying keys on the command-line
  - "Free" specific keys, pinning all other ones, on the command-line
  - Applying weights, as above
  - Simulated annealing with various accept-worse functions, ie: always, drop-slow, linear, drop-fast, never
- Documented utility and code 

TODO- Short bullet list of main features (CLI commands, API functions, supported algorithms, integrations, etc.)
- Performance characteristics, supported platforms (Linux, macOS, Windows)
- Compatibility notes (Go version, module path)

## Supported Metrics
The following metrics are currently supported by Keycraft. Spaces in the corpus are discarded. Keycraft aims to follow the Keyboard Layouts Doc (KLD). Examples are on the Qwerty layout, which is obviously row-staggered.

### Bigram metrics
- SFB (Same Finger Bigram) - percentage of bigrams typed using the same finger (excluding identical-key repeats), for example "ed" but not "ee".
- LSB (Lateral Stretch Bigram) - percentage of bigrams that map to pre-defined lateral-stretch finger pairs, as described in KLB, for example "te" or "be".
- FSB (Full Scissor Bigram) - percentage of bigrams forming pre-defined full-scissor patterns (>1.5U vertical separation), as described in KLB, for example "ct".
- HSB (Half Scissor Bigram) - percentage of bigrams forming pre-defined half-scissor patterns (<=1.5U vertical separation), as described in KLB, for example "st".

### Skipgram metrics
- SFS (Same Finger Skipgram) - percentage of skipgrams typed using the same finger (excluding identical-key repeats), for example "tor".
- LSS (Lateral Stretch Skipgram) - percentage of skipgrams that map to lateral-stretch pairs, for example "the" or "ble".
- FSS (Full Scissor Skipgram) - percentage of skipgrams forming full-scissor patterns, for example "roc".
- HSS (Half Scissor Skipgram) - percentage of skipgrams forming half-scissor patterns, for example "rus".

### Trigram metrics

### Alternations - First key on one hand, the second key on the other, the last key on the first hand again
- ALT (Alternation total) - total percentage of hand alternations (ALT-OTH + ALT-SFS).
- ALT-SFS (Alternation — Same Finger Skipgram) - portion of cross-hand trigram alternations that are same‑finger alternations (excluding identical-key skips), for example "and" and "ent", but not "ana" nor "ene".
- ALT-OTH (Alternation — Other) - portion of cross-hand trigram alternations not classified as SFS (normal alts).

### Two-rolls - Two keys on one hand (rolling in- or outward), and one on the other
- 2RL (Two-key Rolls total) - total percentage for two-key in- and out-rolls (2RL-IN + 2RL-OUT).
- 2RL-IN (Two-key Rolls — Inward) - two-key roll trigrams classified as inward rolls, for example "ing".
- 2RL-OUT (Two-key Rolls — Outward) - two-key roll trigrams classified as outward rolls, for example "tio".
- 2RL-SFB (Two-key Rolls — Same-Finger) - two-key roll trigrams where both keys use the same finger, for example "all". Argueably, identical-key SF are not inconvenient and should be separated out.

### Three-rolls - All three keys are typed on one hand
- 3RL (Three-key Rolls total) - total percentage for three-key in- and out-rolls (3RL-IN + 3RL-OUT).
- 3RL-IN (Three-key Rolls — Inward) - three-key roll trigrams classified as inward sequences, for example "act".
- 3RL-OUT (Three-key Rolls — Outward) - three-key roll trigrams classified as outward sequences, for example "rea".
- 3RL-SFS (Three-key Rolls — Same Finger Skipgram) - three-key roll trigrams with same-finger involvement, for example "ted" or "ill". Argueably, identical-key SFS are not inconvenient and should be separated out.

### Redirections - All three keys on one hand
- RED (Redirections total) - total percentage of redirections (sum of RED-* categories).
- RED-WEAK (Redirections — Weak) - all redirections on one hand with no index involvement (weaker redirections), for example "was".
- RED-SFS (Redirections — Same Finger Skipgram) - redirections that involve same finger skipgrams (excluding identical-key repeats), for example "you".
- RED-OTH (Redirections — Other) - other (normal) redirection types on one hand, for example "ion".

### Other metrics
- IN:OUT (Inward:Outward ratio) - ratio of inward rolls to outward rolls computed as (2RL-IN + 3RL-IN) / (2RL-OUT + 3RL-OUT).
- FBL (Finger BaLance) - cumulative absolute deviation (percentage points) from the ideal finger-load distribution.
- POH (Pinky Off Home) - percentage of unigram frequency typed with a pinky while that pinky is off its home row.

#### Ideal Finger-load distribution
|               | Left-Pinky | Left-Ring | Left-Middle | Left-Index | Right-Index | Right-Middle | Right-Ring | Right-Pinky |
|--------------:|:----------:|:---------:|:-----------:|:----------:|:-----------:|:------------:|:----------:|:-----------:|
| Ideal load (%)| 8.0        | 11.0      | 16.0        | 15.0       | 15.0        | 16.0         | 11.0       | 8.0         |

### Hand balance metrics
This does not include space because n-grams with spaces are discarded from the corpus.

- H0, H1 (Hand usage) - percentage of total keystrokes by each hand (H0 = left, H1 = right). 
- F0–F9 (Finger usage) - percentage of total keystrokes by each finger (F0 = left pinky … F4 = left thumb, F5 = right thumb … F9 = right pinky).
- C0–C11 (Column usage) - percentage of total keystrokes per physical column on the layout.
- R0–R3 (Row usage) - percentage of total keystrokes per physical row on the layout.

## Installation TO DO
- Prebuilt binaries (if available): link or note where to find them.
- From source:
  ```bash
  go mod download
  go build ./...
  ```
- To install via `go install`:
  ```bash
  go install github.com/rbscholtus/keycraft/cmd/keycraft@latest
  ```

## Usage

### Getting help

Use the `help` command to get help for the tool or a specific command, for example:

```bash
keycraft help
keycraft help optimise
keycraft h o
```

- Most commands and flags have shortened versions to avoid typing too much.

### Viewing one or more layouts

Use the `view` command and specify the layout(s) you want to view. The layouts you specify must be files that are located in the `./data/layouts` directory. To view your own layout, add the `.klf` file for your layout here first.

```bash
keycraft view focal.klf gallium-v2.klf
```

The corpus that is used to generate the stats is `./data/corpus/default.txt`. At the moment this is Shai's Cleaned iweb (90m words), available from:
  https://colemak.com/pub/corpus/iweb-corpus-samples-cleaned.txt.xz

To change to any of the corpuses in `./data/corpus`, use the `corpus` command:

```bash
keycraft view -c monkeyracer.txt focal.klf gallium-v2.klf
```

The first time a corpus is used (or after a corpus has changed), a cache is generated that will make loading it faster next time.

### Analysing and comparing one or more layouts

Use the `analyse` command and specify the layout(s) you want to analyse.

```bash
keycraft analyse focal.klf sturdy.klf gallium-v2.klf
```

Note that the trigrams totals listed in the tables will be higher than the stats in the overview. This is because the tables show more than what is included in the overview. That's why they're called __detailed__ tables!

### Ranking layouts

Use the `rank` command to rank and compare a large number of layouts. Layouts are ranked by their overall score which depends on the weights you assign to each of the metrics, as well as the corpus you use. The weights that are applied are shown in the table's header.

```bash
keycraft rank  ## to rank all layouts in ./data/layouts
keycraft rank --deltas rows  ## show the difference between each pair of rows, to easily compare layouts
keycraft rank --deltas median  ## add a median layout (ranked #0), and show the difference between this median layout and the other layouts
keycraft rank --deltas canary.klf  ## show the difference between canary.klf and the other layouts
keycraft rank --metrics extended  ## show more columns with more metrics
keycraft rank --weights sfb=-1000  ## override the weight of the SFB metric with a high value (effectively making the ranks SFB-based ranks)
keycraft rank -d canary.klf colemak.klf "colemak-qi;x.klf" colemak-dh.klf  ## show only a few layouts, comparing them against canary.klf
```

- Better layouts appear at the top of the list. `qwerty.klf` appears at the bottom of the list!
- The median layout is determined by taking the median of all layouts for each metric, normalising all metrics, and calculating the median layout's score by applying weights.
- Default weights are specified in the file `./data/weighs/default.txt`. You can either specify a different weights file using the `--weights-file` flag, or override specific weights using the `--weights` flag.
- The metrics shown in the table do not affect the calculation of each layout's score. The basic table for example shows a weight of 0 for ALT, which makes you think Alts do not affect the scores. But actually if you show extended metrics, you will see a weight is applied to ALT-SFS.

### Optimising a layout

Use the `optimise` command and specify the layout you want to optimise.

```bash
keycraft optimise qwerty.klf
```

## Configuration

### Specifying and choosing a suitable corpus (for all commands)

TO DO

### Specifying weights (for ranking and optimising)

- Describe config locations, file format (YAML/JSON), and common options.
	layoutDir  = "data/layouts/"
	corpusDir  = "data/corpus/"
	weightsDir = "data/weights/"
	pinsDir    = "data/pins/"

## Contributing
- Questions, suggestions, and feedback are super welcome! Just open a New Issue and I'll get back to you as soon as I can.
- I probably can't take PRs until I feel a solid base implementation is in place.

## License
BSD-3-Clause license. See LICENSE file for details.

## Contact
- Author: Barend Scholtus <barend.scholtus@gmail.com>
- Issue tracker: https://github.com/rbscholtus/keycraft/issues

## Planned features
The below features are needed before I would even consider calling it Keycraft 1.0:
- Add detailed tables with metric-specific colunms - GOING WELL
- Split SF in same index <> other index?
- Add option to specify the number of rows in the tables
- Add Alt fingering de-buff
- Refactor I/O <> internal GOING WELL just do ranking.go
- Add Norvig corpus, reddit-small corpus
- Add metrics based on Dist vs %
- Add Effort stat (like cyano?)
- Add Hard words (like cyano?)
- Testing and documentation of all analysis
- Concurrent loading of layouts
- Put Optimise in its own struct
- Optimise optimisation

The below features will hopefully start to add some value:
- Add support for more than 1 layer
- support angle mod (?)
- investigate space analysis

Note: the main package is located in `./cmd/keycraft`.
