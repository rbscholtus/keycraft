# Keycraft

Keycraft is a command-line utility for analying and optimising keyboard layouts. It's goal is to help the user design their "ideal" layout.

## Table of contents
- Features
- Quick start
- Installation
- Usage
- Configuration
- Development
- Testing
- Contributing
- License
- Contact

## Overview
Keycraft is a command-line utility for analying and optimising keyboard layouts. It's goal is to help the user design their "ideal" layout.

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
This does not include space because n-grams with space are discarded from the corpus.

- H0, H1 (Hand usage) - percentage of total keystrokes by each hand (H0 = left, H1 = right). 
- F0–F9 (Finger usage) - percentage of total keystrokes by each finger (F0 = left pinky … F4 = left thumb, F5 = right thumb … F9 = right pinky).
- C0–C11 (Column usage) - percentage of total keystrokes per physical column on the layout.
- R0–R3 (Row usage) - percentage of total keystrokes per physical row on the layout.

## Quick start
1. Install (if CLI):
   ```bash
   go install github.com/yourusername/keycraft@latest
   ```
2. Build from source:
   ```bash
   git clone https://github.com/yourusername/keycraft.git
   cd keycraft
   go build ./...
   ```
3. Run a simple example:
   ```bash
   # CLI example
   keycraft generate --name mykey

   # or as a library
   go run ./examples/simple
   ```

## Installation
- Prebuilt binaries (if available): link or note where to find them.
- From source:
  ```bash
  go mod download
  go build ./...
  ```

## Usage
- CLI: brief command list with examples
  ```bash
  keycraft help
  keycraft generate --type rsa --bits 2048
  keycraft export --format pem --out key.pem
  ```
- Library: short example of importing and using a primary package
  ```go
  import "github.com/yourusername/keycraft/pkg/keycraft"

  // ...example code...
  ```
- Note about configuration files / environment variables (if any)

## Configuration
- Describe config locations, file format (YAML/JSON), and common options.
- Example config snippet.

## Development
- Prerequisites: Go version (e.g., Go 1.20+), other tools.
- How to run linters, formatters:
  ```bash
  go vet ./...
  go fmt ./...
  golangci-lint run
  ```
- Project layout overview (brief) — main packages and responsibilities.

## Testing
- Run unit tests:
  ```bash
  go test ./...
  ```
- How to run integration tests or examples.

## Contributing
- Short contribution workflow: fork, branch, PR, code review.
- Coding conventions and commit message style.
- How to run pre-commit checks locally.

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
