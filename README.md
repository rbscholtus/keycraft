# Keycraft

![Go Version](https://img.shields.io/github/go-mod/go-version/rbscholtus/keycraft)
[![License](https://img.shields.io/github/license/rbscholtus/keycraft)](./LICENSE)
[![Release](https://img.shields.io/github/v/release/rbscholtus/keycraft)](https://github.com/rbscholtus/keycraft/releases)
[![Build](https://github.com/rbscholtus/keycraft/actions/workflows/go.yml/badge.svg)](https://github.com/rbscholtus/keycraft/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/rbscholtus/keycraft.svg)](https://pkg.go.dev/github.com/rbscholtus/keycraft)

Keycraft is a Golang-based command-line utility for analysing, comparing, and optimising keyboard layouts. It helps layout designers quickly evaluate efficiency with detailed metrics, rankings, and visualizations.

## Basic analysis example (QWERTY; Shai corpus)

```
╭     ┬                                                       ╮
                                QWERTY                         
├     ┼                                                       ┤
 Board ╭───┬───┬───┬───┬───┬───╮  ╭───┬───┬───┬───┬───┬───╮    
       │   │ q │ w │ e │ r │ t │  │ y │ u │ i │ o │ p │ \ │    
       ╰┬──┴┬──┴┬──┴┬──┴┬──┴┬──┴╮ ╰┬──┴┬──┴┬──┴┬──┴┬──┴┬──┴╮   
        │   │ a │ s │ d │ f │ g │  │ h │ j │ k │ l │ ; │ ' │   
        ╰─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─╮╰─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─╮ 
          │   │ z │ x │ c │ v │ b │  │ n │ m │ , │ . │ / │   │ 
          ╰───┴───┴───┼───┼───┼───┤  ├───┼───┼───┼───┴───┴───╯ 
                      │   │   │   │  │ _ │   │   │             
                      ╰───┴───┴───╯  ╰───┴───┴───╯             
├     ┼                                                       ┤
 Hand   ╭───────┬───┬────┬────────┬────────┬───┬────┬───────╮  
        │   LP  │ LR│ LM │   LI   │   RI   │ RM│ RR │   RP  │  
        ├───┬───┼───┼────┼───┬────┼────┬───┼───┼────┼───┬───┤  
        │0.0│8.2│8.5│18.5│9.3│12.5│13.3│5.5│9.0│12.9│2.2│0.3│  
        ├───┴───┼───┼────┼───┴────┼────┴───┼───┼────┼───┴───┤  
        │  8.2  │8.5│18.5│  21.8  │  18.8  │9.0│12.9│  2.4  │  
        ├───────┴───┴────┴────────┼────────┴───┴────┴───────┤  
        │          56.9%          │          43.1%          │  
        ╰─────────────────────────┴─────────────────────────╯  
├     ┼                                                       ┤
 Row              ╭───────┬───────┬────────┬───────╮           
                  │  Top  │  Home │ Bottom │ Thumb │           
                  ├───────┼───────┼────────┼───────┤           
                  │ 50.4% │ 32.2% │  17.4% │  0.0% │           
                  ╰───────┴───────┴────────┴───────╯           
├     ┼                                                       ┤
 Stats   ╭───────────┬────────────┬────────────┬───────────╮   
         │SFB:  6.52%│LSB:   3.36%│FSB:   1.19%│HSB:  4.75%│   
         ├───────────┼────────────┼────────────┼───────────┤   
         │SFS: 11.08%│LSS:   6.50%│FSS:   1.57%│HSS:  5.19%│   
         ├───────────┼────────────┼────────────┼───────────┤   
         │RED: 13.04%│.NML:  7.38%│.WEAK: 1.41%│.SFS: 4.25%│   
         ├───────────┼────────────┼────────────┼───────────┤   
         │ALT: 26.11%│.NML: 20.65%│.SFS:  5.45%│           │   
         ├───────────┼────────────┼────────────┼───────────┤   
         │2RL: 46.42%│.IN:  19.85%│.OUT: 16.97%│.SFB: 9.60%│   
         ├───────────┼────────────┼────────────┼───────────┤   
         │3RL: 11.82%│.IN:   1.29%│.OUT:  1.45%│.SFB: 9.08%│   
         ├───────────┼────────────┼────────────┼───────────┤   
         │FLW: 60.21%│I:O:    1.15│            │           │
         ├───────────┼────────────┼────────────┼───────────┤
         │HLD:   0.00│FLD:   0.00%│RLD:   0.00%│POH:  4.80%│
         ╰───────────┴────────────┴────────────┴───────────╯   
╰     ┴                                                       ╯
```

## Quick Start

### Installing Keycraft

For those that do not have Golang installed on their system:

- Download the [latest Keycraft binary](https://github.com/rbscholtus/keycraft/releases) for your system.
- Additionally, download the `data.tar.gz` archive.
- Create a new directory on your system, let's say `Downloads/keycraft`.
- Move the binary you downloaded to that directory.
- Extract the files in `data.tar.gz` to that directory.

Now you should have something like:

```
.
├── data
│   ├── corpus
│   │   ├── akl.txt.json
│   │   ├── default.txt.json
│   │   ├── monkeyracer.txt.json
│   │   └── shai.txt.json
│   ├── layouts
│   │   ├── ...
│   │   └── ...
│   ├── config
│   │   ├── focal.pin
│   │   ├── load_targets.txt
│   │   ├── qwerty.pin
│   │   └── weights.txt
└── keycraft-darwin-arm64
```

- Open the new directory in a terminal window.
- Rename your binary to `keycraft` for ease of use.
- On Mac/Linux, make the `keycraft` file executable:
  - `chmod +x keycraft`
- Run keycraft, for example on Mac/Linux:
  - `./keycraft h`
- When the program is blocked by your OS, allow running it. For example on Mac, go to Privacy and Security in the System Settings, and allow the program.

Now this should run! Proceed with the quick examples below.

```bash
# View corpus statistics for the default corpus (Shai)
keycraft c

# View a built-in layout
keycraft v qwerty

# Analyse and compare multiple layouts
keycraft a qwerty colemak dvorak

# Rank built-in layouts with custom weights (making SFB the dominant metric)
# Note many metrics are considered negative; they are written with a -
keycraft r --weights sfb=-1000

# Optimise the qwerty layout but keeping its signature keys in place
keycraft o --pins qwerty qwerty
```

### Installing and running Keycraft with Golang installed on your system

You should be able to install Keycraft with the below command:

```bash
# Install
go install github.com/rbscholtus/keycraft/cmd/keycraft@latest
```

## Features

### Core Features

- View corpus statistics
- View layouts, with 140+ built-in layouts
- Tabulate hand, finger, row, and column usage stats
- Compare layouts side by side
- Analyse detailed layout metrics in tables
- Rank layouts using customizable weights
- Optimise layouts

### Advanced Features

- Supports 4x6 + 3 (x2) layouts (row-staggered, angle-modded, ortholinear, and column-staggered)
- Supports 29 layout metrics and 28 counts
- Supports Euclidian distance specific to each physical layout type
- Supports MonkeyRacer, Shai (default), and AKL corpus files out of the box
- Supports a default corpus, eliminating the need to specify the corpus for every command
- Supports an internal cache for fast loading of corpuses
- Supports viewing corpus statistics (word length frequency, top n-gram, top words)
- Supports scoring and ranking some or all layouts
- Supports showing deltas between rows for comparing how metrics differ between layouts
- Supports showing deltas between each layout and the median, or a specific base layout
- Supports user-defined weights for all metrics, using a weights file and from the command-line
- Supports a default weights file, eliminating the need to specify the weights for every command
- Supports normalisation of metrics using stable scaling (using the median and inter-quartile range of each metric)
- Supports optimising a layout using Breakout Local Search (BLS)
- Supports pinning specific keys using a .pin file and from the command-line
- Supports "freeing" specific keys (pinning all others) from the command-line
- Supports MacOS (tested), Linux, Windows (tested)
- Supports documentation LOL

## Supported Metrics

Keycraft supports the following metrics. Here are some notes:

- Bigram, skipgram, and trigram metrics follow the Keyboard Layouts Doc.
- Examples are based on the Qwerty layout.
- Spaces in the corpus are discarded.

### Metrics

#### Bigram Metrics
| Acronym  | Metric                              | Description                                                           | Examples            |
|----------|-------------------------------------|-----------------------------------------------------------------------|---------------------|
| SFB      | Same Finger Bigram                  | Percentage of bigrams typed using the same finger (excluding repeats) | "ed", "lo" (not "ee") |
| LSB      | Lateral Stretch Bigram              | Percentage of bigrams that map to lateral-stretch finger pairs        | "te", "be"          |
| FSB      | Full Scissor Bigram                 | Percentage of bigrams forming scissor patterns that skip the home row    | "ct", "ex"          |
| HSB      | Half Scissor Bigram                 | Percentage of bigrams forming scissor patterns that involve the home row | "st", "ca"          |

#### Skipgram Metrics
| Acronym  | Metric                              | Description                                                           | Examples            |
|----------|-------------------------------------|-----------------------------------------------------------------------|---------------------|
| SFS      | Same Finger Skipgram                | Percentage of skipgrams typed using the same finger (excluding repeats) | "end", "tor" (not "ene") |
| LSS      | Lateral Stretch Skipgram            | Percentage of skipgrams that map to lateral-stretch pairs             | "the", "ble"        |
| FSS      | Full Scissor Skipgram               | Percentage of skipgrams forming full-scissor patterns                 | "cut", "roc"        |
| HSS      | Half Scissor Skipgram               | Percentage of skipgrams forming half-scissor patterns                 | "sit", "rus"        |

#### Trigram Metrics
| Acronym  | Metric                              | Description                                                    | Examples            |
|----------|-------------------------------------|----------------------------------------------------------------|---------------------|
| RED      | Redirections total                  | Total % of redirections                                        |                     |
| RED-WEAK | Redirections — Weak                 | Redirections on one hand with no index and thumb involvement             | "was", "ese"        |
| RED-SFS  | Redirections — Same Finger Skipgram | Redirections on one hand that are same-finger skipgrams        | "you", "ter"        |
| RED-NML  | Redirections — Other                | Other (normal) redirections on one hand                        | "ion", "ate", "ere" |
| ALT      | Alternation total                   | Total % of hand alternations (ALT-NML + ALT-SFS)               |                     |
| ALT-SFS  | Alternation — Same Finger Skipgram  | Cross-hand alternations that are same-finger alternations      | "for", "men"        |
| ALT-NML  | Alternation — Normal                | Cross-hand alternations not classified as SFS     | "and", "ent", "iti" |
| 2RL      | 2-key Rolls total                   | Total % of two-key roll trigrams (2RL-IN + 2RL-OUT + 2RL-SFB)  |                     |
| 2RL-IN   | 2-key Rolls — Inward                | Two-key roll trigrams classified as inward rolls               | "ing", "hat"        |
| 2RL-OUT  | 2-key Rolls — Outward               | Two-key roll trigrams classified as outward rolls              | "tio", "thi"        |
| 2RL-SFB  | 2-key Rolls — Same Finger Bigram    | Two-key rolls where both keys use the same finger              | "nce", "all"        |
| 3RL      | 3-key Rolls total                   | Total % of three-key roll trigrams (3RL-IN + 3RL-OUT + 3RL-SFB) |                     |
| 3RL-IN   | 3-key Rolls — Inward                | Three-key roll trigrams classified as inward sequences         | "act", "lin"        |
| 3RL-OUT  | 3-key Rolls — Outward               | Three-key roll trigrams classified as outward sequences        | "rea", "tes"        |
| 3RL-SFB  | 3-key Rolls — Same Finger Bigram    | Three-key rolls where first and last keys use the same finger  | "ted", "ill"        |

#### Flow Metrics
| Acronym  | Metric                              | Description                                                    | Examples            |
|----------|-------------------------------------|----------------------------------------------------------------|---------------------|
| FLW      | Flowiness                           | Flow measure: ALT-NML + 2RL-IN + 2RL-OUT + 3RL-IN + 3RL-OUT   |                     |
| IN:OUT   | Inward:Outward rolls ratio          | Ratio of inward to outward rolls: (2RL-IN + 3RL-IN) / (2RL-OUT + 3RL-OUT) |                     |

#### Load Distribution Deviation & Penalty Metrics
| Acronym  | Metric                              | Description                                                    | Examples            |
|----------|-------------------------------------|----------------------------------------------------------------|---------------------|
| HLD      | Hand Load Deviation                 | Deviation from target hand load distribution (see below)            |                     |
| FLD      | Finger Load Deviation               | Deviation from target finger load distribution (see below)          |                     |
| RLD      | Row Load Deviation                  | Deviation from target row load distribution (see below)             |                     |
| POH      | Pinky Off Home (Weighted)           | Weighted penalty for off-home pinky usage (see below)          |                     |

#### Usage Distribution Measures

These measures report actual keystroke percentages. Unlike HLD/FLD/RLD, these are raw measurements, not deviations from targets.

| Acronym  | Measure                             | Description                                                    | Examples            |
|----------|-------------------------------------|----------------------------------------------------------------|---------------------|
| H0, H1   | Hand usage percentages              | Left/right hand % (main rows 0-2 only)                         |                     |
| F0–F9    | Finger usage percentages            | Per-finger % (main rows 0-2 only). See below for mapping       |                     |
| C0–C11   | Column usage percentages            | Per-column % (main rows 0-2 only)                              |                     |
| R0–R3    | Row usage percentages               | Per-row % (all rows 0-3, includes thumbs)                      |                     |

**Mappings**:
- **F0-F9**: F0=Left Pinky, F1=Left Ring, F2=Left Middle, F3=Left Index, F4=Left Thumb (0% in main row counts), F5=Right Thumb (0% in main row counts), F6=Right Index, F7=Right Middle, F8=Right Ring, F9=Right Pinky
- **R0-R3**: R0=Top row, R1=Home row, R2=Bottom row, R3=Thumb row

### Load Deviation & Penalty Metrics

Here are the detailed descriptions of the above listed metrics.

- **HLD - Hand Load Deviation**: Measures the cumulative absolute deviation (in percentage points) from the target hand load distribution. Calculated as: HLD = |H0 - target_H0| + |H1 - target_H1|, where H0 and H1 are the actual left/right hand usage percentages. Lower values indicate better balance. Only counts main rows (0-2), excluding thumb cluster.

- **FLD - Finger Load Deviation**: Measures the cumulative absolute deviation (in percentage points) from the target finger load distribution across 8 fingers (F0-F3, F6-F9). Calculated as: FLD = Σ|Fi - target_Fi| for all non-pinky fingers, plus only positive deviations for pinkies (F0, F9). The asymmetric calculation for pinkies avoids penalizing layouts that successfully reduce pinky load below target. Lower values indicate better balance. Only counts main rows (0-2), excluding thumb cluster.

FLD includes only the main finger rows. What this means is that if an alpha key is moved to the thumb cluster and no other changes are made, each finger (except the one that used to type the moved key), will now have a higher load relative to the total load on all 8 fingers, and FLD wil go up.

- **RLD - Row Load Deviation**: Measures the weighted deviation from the target row load distribution across three main rows (top, home, bottom). Uses directional penalties: home row penalizes below-target usage (encouraging home row), while top/bottom rows penalize above-target usage (discouraging those rows). Calculated as: RLD = -(home_actual - home_target) + (top_actual - top_target) + (bottom_actual - bottom_target). Lower values indicate better balance. Only counts main rows (0-2), excluding thumb cluster.

- **POH - Pinky Off Home**: A weighted penalty score for pinky key usage, focusing on positions outside the ideal home row spot to minimize strain on the weakest finger. Each pinky position has a configurable penalty weight, with higher values indicating greater discomfort or penalty. Calculates as: the sum of (key frequency × position weight) for all pinky keys, expressed as a percentage of total keystrokes. Lower values are better.

### Target Definitions

- **Target Hand Load Distribution**: The target distribution of typing load across the two hands, including only the fingers (excluding thumbs). It is configurable, with defaults of left: 50%, right: 50%. Values are normalized to sum to 100%.

- **Target Finger Load Distribution**: The target distribution of typing load across the eight fingers (left and right pinky, ring, middle, and index). It is configurable, with defaults of left pinky: 7%, left ring: 10%, left middle: 16%, left index: 17%, right index: 17%, right middle: 16%, right ring: 10%, right pinky: 7%. Values are normalized to sum to 100%.

- **Target Row Load Distribution**: The target distribution of typing load across the three main rows (top, home, and bottom), excluding the thumb cluster. It is configurable, with defaults of top row: 17.5%, home row: 75.0%, bottom row: 7.5%. Values are normalized to sum to 100%.

- **Pinky Off Home (POH) Weights**: The weights for calculating the Pinky Off Home penalty. Defaults vary by position: 0.0 for home-inner (ideal), 1.0 for home-outer, 1.5 for top/bottom-inner, and 2.0 for top/bottom-outer (mirrored for both hands).

```
Target Loads and Penalty Weights

Hand   ╭─────┬──────┬──────┬──────┬──────┬──────┬──────┬─────╮
       │  LP │  LR  │  LM  │  LI  │  RI  │  RM  │  RR  │  RP │
       ├─────┼──────┼──────┼──────┼──────┼──────┼──────┼─────┤
       │ 7.0 │ 10.0 │ 16.0 │ 17.0 │ 17.0 │ 16.0 │ 10.0 │ 7.0 │
       ├─────┴──────┴──────┴──────┼──────┴──────┴──────┴─────┤
       │           50%            │            50%           │
       ╰──────────────────────────┴──────────────────────────╯  
Row              ╭───────┬───────┬────────┬───────╮           
                 │  Top  │  Home │ Bottom │ Thumb │           
                 ├───────┼───────┼────────┼───────┤           
                 │ 17.5% │ 75.0% │  7.5%  │  N/A  │           
                 ╰───────┴───────┴────────┴───────╯           
Pinky  ╭───┬───┬───┬───┬───┬───╮  ╭───┬───┬───┬───┬───┬───╮    
       │2.0│1.5│   │   │   │   │  │   │   │   │   │1.5│2.0│    
       ╰┬──┴┬──┴┬──┴┬──┴┬──┴┬──┴╮ ╰┬──┴┬──┴┬──┴┬──┴┬──┴┬──┴╮   
        │1.0│ 0 │   │   │   │   │  │   │   │   │   │ 0 │1.0│   
        ╰─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─╮╰─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─╮ 
          │2.0│1.5│   │   │   │   │  │   │   │   │   │1.5│2.0│ 
          ╰───┴───┴───┼───┼───┼───┤  ├───┼───┼───┼───┴───┴───╯ 
                      │   │   │   │  │   │   │   │             
                      ╰───┴───┴───╯  ╰───┴───┴───╯             
```

### Load Distribution Considerations

#### Finger Load Distribution

Finger load distribution in ergonomic keyboard layouts aims to allocate typing effort based on finger strength and dexterity. Stronger, more central fingers (index and middle) should handle higher loads, while weaker ones (pinky and ring) take less to reduce strain and fatigue. This is a core principle in optimizers like Carpalx, which penalizes overuse of weaker fingers through its effort model. Other sources (e.g., Colemak forums, Workman philosophy, and ergonomic studies) converge on similar non-uniform distributions, often assuming symmetry between hands.

Here's a synthesis of reasonable targets:

| Finger (Per Hand) | Reasonable Load Range (%) | Rationale / Sources |
|----------|---------------------------------|-----------------------------|
| Pinky | 6–8% | Weakest finger; minimize to avoid strain. Carpalx penalizes pinky heavily; Workman and Colemak aim low here. Ergonomic reviews (e.g., NIH studies on wrist deviation) note pinky overuse contributes to RSI. |
| Ring | 8–12% | Slightly stronger than pinky but still limited; sources like Hands Down and MTGAP layouts target this to balance with middle finger. |
| Middle | 12–15% | Strong and central; can handle moderate-high load. Colemak and Carpalx variants often place vowels here for efficiency. |
| Index | 15–20% | Most dexterous; handles higher load but not overload (to avoid hand displacement). Workman reduces index stretching vs. Colemak; PDF study on English layouts weights index heavily for common bigrams. |

#### Row Load Distribution

Row distribution prioritizes the home row for most typing, as it aligns with natural finger rest positions and minimizes vertical reach (reducing extension/flexion strain). Carpalx incorporates row penalties in its model—higher effort for top/bottom rows due to distance from home. Optimized Carpalx layouts (e.g., QGMLWB) achieve ~70% home row usage.

Common targets from other analyzers and layouts:

| Row | Reasonable Load Range (%) | Rationale / Sources |
|----------|---------------------------------|-----------------------------|
| Top | 15–25% | Requires upward extension. Colemak reduces top-row load vs. QWERTY; ergonomic guides (e.g., Dygma, Truly Ergonomic) penalize it heavily. Kvikk layout: ~15–20%. |
| Home | 60–75% | Core for efficiency; most common letters here. Colemak: 74%; Dvorak: 70%; QWERTY: only 32% (poor). Hands Down and MTGAP aim for 70%+ to keep fingers "fixed." |
| Bottom | 10–15% | The Top row is generally preferred over the Bottom row. Reaching "up" is anatomically easier for most typists than curling the fingers "down" and "in." |

## Usage

### Getting help

Use the `help` command to get help for the tool or a specific command. Most commands and flags have short versions to avoid typing too much.

```bash
# Get help on available commands
keycraft help

# Get help on a specific commands
keycraft help optimise

# Short version of the above
keycraft h o
```

Example output:

```bash
NAME:
   keycraft - A CLI tool for crafting better keyboard layouts

USAGE:
   keycraft [global options] command [command options]

COMMANDS:
   corpus, c    Display statistics for a text corpus
   view, v      Analyse and display one or more keyboard layouts
   analyse, a   Analyse one or more keyboard layouts in detail
   rank, r      Rank keyboard layouts and optionally view deltas
   optimise, o  Optimise a keyboard layout using Breakout Local Search (BLS)
   help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help
```

### Viewing one or more layouts

Use the `view` command and specify the layout(s) you want to view.

```bash
# View 1 or more layouts side-by-side
keycraft v focal gallium-v2

# View layouts with metrics based on another corpus
keycraft v -c monkeyracer.txt focal gallium-v2
```

- The layouts must be located in `./data/layouts`. To view your own layout, add the `.klf` file for your layout there.
- The corpus that is used to generate the stats is `./data/corpus/default.txt`. At the moment this is Shai's Cleaned iweb (90m words), available from:
  <https://colemak.com/pub/corpus/iweb-corpus-samples-cleaned.txt.xz>
- The first time a corpus is used (or after a corpus has changed), a cache is generated that will make loading it a lot faster next time.

### Analysing and comparing one or more layouts

Use the `analyse` command and specify the layout(s) you want to analyse.

```bash
# Analyse multiple layouts with detailed tables for each metric
keycraft a focal sturdy
```

```
╭     ┬                                                      ┬                                                       ╮
                                FOCAL                                                  STURDY                         
├     ┼                                                      ┼                                                       ┤
 Board          ╭───┬───┬───╮          ╭───┬───┬───╮          ╭───┬───┬───┬───┬───┬───╮  ╭───┬───┬───┬───┬───┬───╮    
        ╭───┬───┤ l │ h │ g ├───╮  ╭───┤ f │ o │ u ├───┬───╮  │   │ v │ m │ l │ c │ p │  │ x │ f │ o │ u │ j │   │    
        │   │ v ├───┼───┼───┤ k │  │ q ├───┼───┼───┤ j │   │  ╰┬──┴┬──┴┬──┴┬──┴┬──┴┬──┴╮ ╰┬──┴┬──┴┬──┴┬──┴┬──┴┬──┴╮   
        ├───┼───┤ r │ n │ t ├───┤  ├───┤ c │ a │ e ├───┼───┤   │   │ s │ t │ r │ d │ y │  │ . │ n │ a │ e │ i │   │   
        │   │ s ├───┼───┼───┤ b │  │ y ├───┼───┼───┤ i │ / │   ╰─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─╮╰─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─╮ 
        ├───┼───┤ x │ m │ d ├───┤  ├───┤ w │ . │ ; ├───┼───┤     │   │ z │ k │ q │ g │ w │  │ b │ h │ ' │ ; │ , │   │ 
        │   │ z ├───┼───┼───┤ p │  │ ' ├───┼───┼───┤ , │   │     ╰───┴───┴───┼───┼───┼───┤  ├───┼───┼───┼───┴───┴───╯ 
        ╰───┴───╯   │   │   ├───┤  ├───┤   │   │   ╰───┴───╯                 │   │   │   │  │ _ │   │   │             
                    ╰───┴───┤   │  │ _ ├───┴───╯                             ╰───┴───┴───╯  ╰───┴───┴───╯             
                            ╰───╯  ╰───╯                                                                              
...
 SFB         ╭───┬───┬─────────┬─────┬────┬──┬───┬────╮              ╭──┬───┬─────────┬─────┬────┬──┬───┬────╮        
             │   │SFB│    COUNT│    %│DIST│HD│FGR│ΔROW│              │  │SFB│    COUNT│    %│DIST│HD│FGR│ΔROW│        
             ├───┼───┼─────────┼─────┼────┼──┼───┼────┤              ├──┼───┼─────────┼─────┼────┼──┼───┼────┤        
             │  1│ue │  440,608│0.13%│1.00│ 2│  9│1.00│              │ 1│ue │  440,608│0.13%│1.03│ 2│  9│1.00│        
             │  2│pt │  267,309│0.08%│1.49│ 1│  4│1.10│              │ 2│n. │  321,343│0.09%│1.00│ 2│  7│0.00│        
             │  3│rl │  261,012│0.08%│1.00│ 1│  2│1.00│              │ 3│rl │  261,012│0.08%│1.03│ 1│  3│1.00│        
             │  4│oa │  254,172│0.07%│1.00│ 2│  8│1.00│              │ 4│oa │  254,172│0.07%│1.03│ 2│  8│1.00│        
             │  5│cy │  110,370│0.03%│1.00│ 2│  7│0.10│              │ 5│nf │  206,315│0.06%│1.03│ 2│  7│1.00│        
             │  6│dg │   89,160│0.03%│2.00│ 1│  4│2.00│              │ 6│dy │  168,246│0.05%│1.00│ 1│  4│0.00│        
             │  7│a. │   82,362│0.02%│1.00│ 2│  8│1.00│              │ 7│cy │  110,370│0.03%│1.60│ 1│  4│1.00│        
             │  8│o. │   80,239│0.02%│2.00│ 2│  8│2.00│              │ 8│tm │   95,503│0.03%│1.03│ 1│  2│1.00│        
             │  9│hn │   78,587│0.02%│1.00│ 1│  3│1.00│              │ 9│h. │   90,645│0.03%│1.80│ 2│  7│1.00│        
             │ 10│eu │   68,749│0.02%│1.00│ 2│  9│1.00│              │10│dg │   89,160│0.03%│1.12│ 1│  4│1.00│        
             ├───┼───┼─────────┼─────┼────┼──┼───┼────┤              ├──┼───┼─────────┼─────┼────┼──┼───┼────┤        
             │   │   │2,598,169│0.75%│    │  │   │    │              │  │   │3,101,224│0.90%│    │  │   │    │        
             ╰───┴───┴─────────┴─────┴────┴──┴───┴────╯              ╰──┴───┴─────────┴─────┴────┴──┴───┴────╯        
...
 2RL           ╭─────┬───┬───────────┬──────┬────┬───╮                ╭─────┬───┬───────────┬──────┬────┬───╮         
               │     │2RL│      COUNT│     %│DIST│DIR│                │     │2RL│      COUNT│     %│DIST│DIR│         
               ├─────┼───┼───────────┼──────┼────┼───┤                ├─────┼───┼───────────┼──────┼────┼───┤         
               │    1│the│  6,802,477│ 2.63%│0.00│OUT│                │    1│the│  6,802,477│ 2.63%│0.00│OUT│         
               │    2│ing│  3,233,466│ 1.25%│0.00│IN │                │    2│ing│  3,233,466│ 1.25%│0.00│IN │         
               │    3│and│  3,084,446│ 1.19%│0.00│IN │                │    3│and│  3,084,446│ 1.19%│0.00│IN │         
               │    4│ion│  1,720,124│ 0.67%│0.00│IN │                │    4│ent│  1,528,636│ 0.59%│0.00│IN │         
               │    5│ent│  1,528,636│ 0.59%│0.00│IN │                │    5│for│  1,468,675│ 0.57%│0.00│OUT│         
               │    6│for│  1,468,675│ 0.57%│0.00│OUT│                │    6│you│  1,424,616│ 0.55%│0.00│OUT│         
               │    7│tio│  1,380,462│ 0.53%│0.00│IN │                │    7│tio│  1,380,462│ 0.53%│0.00│IN │         
               │    8│tha│  1,189,955│ 0.46%│0.00│OUT│                │    8│hat│  1,232,975│ 0.48%│0.00│OUT│         
               │    9│all│  1,015,889│ 0.39%│0.00│SFB│                │    9│tha│  1,189,955│ 0.46%│0.00│OUT│         
               │   10│thi│    936,782│ 0.36%│0.00│OUT│                │   10│her│  1,158,203│ 0.45%│0.00│OUT│         
               ├─────┼───┼───────────┼──────┼────┼───┤                ├─────┼───┼───────────┼──────┼────┼───┤         
               │     │   │132,151,717│51.15%│    │   │                │     │   │136,696,023│52.91%│    │   │         
               ╰─────┴───┴───────────┴──────┴────┴───╯                ╰─────┴───┴───────────┴──────┴────┴───╯         
...
```

### Ranking layouts

Use the `rank` command to rank and compare a large number of layouts. Layouts are ranked by their overall score which depends on the weights you assign to each of the metrics, as well as the corpus you use. The weights that are applied are shown in the table's header.

```bash
# Rank all layouts in `./data/layouts`
keycraft r

# Rank specific layouts only
keycraft r canary colemak-dh focal night

# Rank all layouts, showing the deltas to easily compare layouts.
# Keycraft shows the delta between the layouts directly above and below each delta row.
# In this mode, the table is meant to be read from top to bottom. Color is used to indicate a metric got better or worse.
# Deltas have rounding errors occasionally.
keycraft r -d rows

# Rank all layouts, adding a "median" layout, ranked #0. The median layout is based on all layouts in `./data/layouts`.
# Keycraft shows the deltas between the median and the other layouts.
# The delta rows above the median layout show the delta between the median layout and the layout directly above the delta row, and should be read upwards.
# The delta rows below the median layout show the delta between the median layout and the layout directly below the delta row, and should be read downwards.
keycraft r -d median

# Rank all layouts, showing the deltas between `canary` and the other layouts
# Delta rows show the delta between `canary` and the layout directly above or below the delta row (similar to the `-d median` option)
keycraft r -d canary

# Rank specific layouts only, comparing them against `canary`
# Reading the table from row 0 downwards, one observes mostly declining statistics
keycraft r -d canary colemak colemak-qix colemak-dh

# Rank all layouts, showing columns for all metrics
keycraft r -d extended

# Rank all layouts, overriding the weight of the SFB metric
# Specifying a high weight like below will effectively rank layouts based on SFBs only. Note the minus (-) sign!
keycraft r -w sfb=-1000
```

- Better layouts appear at the top of the list. `qwerty` appears at the bottom of the list!
- The median layout is determined by taking the median of all layouts for each metric, normalising all metrics, and calculating the median layout's score by applying weights.
- Default weights are specified in the file `./data/config/weights.txt`. You can either specify a different weights file using the `--weights-file` flag, or override specific weights using the `--weights` flag.

### Optimising a layout

Use the `optimise` command and specify the layout you want to optimise.

```bash
# Optimise a layout with an adjusted number of generations, but default weights
keycraft o -g 500 qwerty

# Optimise an already very good layout with some keys pinned
# Pinning keys prevents those keys from being moved around, which could otherwise ruin the essence of a layout
keycraft o -g 100 --pins srntaeiou focal

# Optimise a layout, strongly aiming for good finger balance, but potentially ruining other metrics
keycraft o -w FBL=-100 -g 100 canary

# Optimise a small number of keys using the --free flag
# Optimising special characters should be used in combination with a more specific corpus
keycraft o -g 50 --free "';,.-/" graphite
```

## Configuration

### Specifying and choosing a suitable corpus (for all commands)

More information will be provided.

### Specifying weights (for ranking and optimising)

- Describe config locations, file format (YAML/JSON), and common options.
 layoutDir  = "data/layouts/"
 corpusDir  = "data/corpus/"
 configDir  = "data/config/"

## Contributing

- Questions, suggestions, and feedback are super welcome! Just open a New Issue and I'll get back to you as soon as I can.
- I probably cannot take PRs until I feel a solid base implementation is in place. Sorry about this!

## License

BSD-3-Clause license. See LICENSE file for details.

## Contact

- Author: Barend Scholtus <barend.scholtus@gmail.com>
- Issue tracker: <https://github.com/rbscholtus/keycraft/issues>
- Discord: @ironcollar
