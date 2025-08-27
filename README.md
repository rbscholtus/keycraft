# Keycraft

![Go Version](https://img.shields.io/github/go-mod/go-version/rbscholtus/keycraft)
[![License](https://img.shields.io/github/license/rbscholtus/keycraft)](./LICENSE)
[![Release](https://img.shields.io/github/v/release/rbscholtus/keycraft)](https://github.com/rbscholtus/keycraft/releases)
[![Build](https://github.com/rbscholtus/keycraft/actions/workflows/go.yml/badge.svg)](https://github.com/rbscholtus/keycraft/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/rbscholtus/keycraft.svg)](https://pkg.go.dev/github.com/rbscholtus/keycraft)

Keycraft is a Golang-based command-line utility for analysing, comparing, and optimising keyboard layouts. It helps layout designers quickly evaluate efficiency with detailed metrics, rankings, and visualizations.

#### Basic analysis example

```
╭     ┬                                                       ╮
                                QWERTY                         
├     ┼                                                       ┤
 Board ╭───┬───┬───┬───┬───┬───╮  ╭───┬───┬───┬───┬───┬───╮    
       │   │ q │ w │ e │ r │ t │  │ y │ u │ i │ o │ p │ \ │    
       ╰┬──┴┬──┴┬──┴┬──┴┬──┴┬──┴╮ ╰┬──┴┬──┴┬──┴┬──┴┬──┴┬──┴╮   
        │   │ a │ s │ d │ f │ g │  │ h │ j │ k │ l │ ; │ ' │   
        ╰─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─╮╰─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─╮ 
          │   │ z │ x │ c │ v │ b │  │ n │ m │ , │ . │ / │   │ 
          ╰───┴───┴───┼───┼───┼───┤  ├───┼───┼───┼───┴───┴───╯ 
                      │   │   │   │  │ _ │   │   │             
                      ╰───┴───┴───╯  ╰───┴───┴───╯             
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
 Stats     ╭───────────┬───────────┬──────────┬───────────╮    
           │SFB:  6.52%│LSB:  3.36%│FSB: 1.19%│HSB:  4.75%│    
           ├───────────┼───────────┼──────────┼───────────┤    
           │SFS: 11.08%│LSS:  6.50%│FSS: 1.57%│HSS:  5.18%│    
           ├───────────┼───────────┼──────────┼───────────┤    
           │ALT: 26.11%│2RL: 36.81%│3RL: 2.74%│RED: 13.04%│    
           ├───────────┼───────────┼──────────┼───────────┤    
           │I:O:   1.15│FBL: 30.35%│POH: 2.33%│WEAK  1.41%│    
           ╰───────────┴───────────┴──────────┴───────────╯    
├     ┼                                                       ┤
```

## Quick Start

### Installing and running with Golang installed on your system

```bash
# Install (requires Golang 1.25 on your system)
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

### Installing without Golang on your system

- Download the latest release for your system from https://github.com/rbscholtus/keycraft/releases
- Download the `data` archive as well.
- Unpack both downloads to a new directory, for example `Downloads/keycraft`
- Open the new directory in a terminal window
- On Mac/Linux, make the `keycraft-xyz` file executable, for example on Mac:
  - `chmod +x keycraft-darwin-arm64`
- Run keycraft, for example on Mac:
  - `./keycraft-darwin-arm64`
- When the program is blocked by your OS, allow running it. For example on Mac, go to Privacy and Security in the System Settings, and allow the program.
- Now this should run! Proceed with the examples above.

## Features

### Core Features

- View layouts, with 30+ built-in layouts
- Visualize hand, finger, row, and column usage
- Compare layouts side by side
- Analyse detailed layout metrics in tables
- Rank layouts using customizable weights
- Optimise layouts with simulated annealing

### Advanced Features

- Supports 4x6+3 (x2) layouts, row-staggered, ortholinear, column-staggered
- Supports Euclidian distance specific to each physical layout type
- Supports MonkeyRacer, Shai, and AKL corpus files out of the box
- Supports an internal cache for fast loading
- Supports a default corpus, eliminating the need to specify the corpus for every command
- Supports scoring and ranking some or all layouts
- Supports showing deltas between rows for comparing how metrics differ between layouts
- Supports showink deltas between each layout and the median or a specified "reference" layout
- Supports user-defined weights for all metrics, using a weights file and from the command-line
- Supports a default weights file, eliminating the need to specify the weights for every command
- Supports normalisation of metrics using stable scaling (using the median and inter-quartile range of each metric)
- Supports optimising a layout using Simulated annealing
- Supports various accept-worse functions: always, drop-slow, linear, drop-fast, never
- Supports pinning specific keys using a .pin file and from the command-line
- Supports "freeing" specific keys (pinning all others) from the command-line
- Supports MacOS (tested), Linux, Windows
- Supports documentation LOL

## Supported Metrics

The following metrics are currently supported by Keycraft. Spaces in the corpus are discarded. Examples are on the Qwerty layout, which is obviously row-staggered and has some stretchgrams that other physical layouts do not have.

| Acronym  | Metric                              | Examples            |
|----------|-------------------------------------|---------------------|
| SFB      | Same Finger Bigram                  | "ed", "lo" (not "ee") |
| LSB      | Lateral Stretch Bigram              | "te", "be"          |
| FSB      | Full Scissor Bigram                 | "ct", "ex"          |
| HSB      | Half Scissor Bigram                 | "st", "ca"          |
| SFS      | Same Finger Skipgram                | "end", "tor" (not "ene") |
| LSS      | Lateral Stretch Skipgram            | "the", "ble"        |
| FSS      | Full Scissor Skipgram               | "cut", "roc"        |
| HSS      | Half Scissor Skipgram               | "sit", "rus"        |
| ALT      | Alternation total                   |                     |
| ALT-SFS  | Alternation — Same Finger Skipgram  | "for", "men"        |
| ALT-OTH  | Alternation — Other                 | "and", "ent", "iti" |
| 2RL      | 2-key Rolls total                   |                     |
| 2RL-IN   | 2-key Rolls — Inward                | "ing", "hat"        |
| 2RL-OUT  | 2-key Rolls — Outward               | "tio", "thi"        |
| 2RL-SFB  | 2-key Rolls — Same Finger Bigram    | "nce", "all"        |
| 3RL      | 3-key Rolls total                   |                     |
| 3RL-IN   | 3-key Rolls — Inward                | "act", "lin"        |
| 3RL-OUT  | 3-key Rolls — Outward               | "rea", "tes"        |
| 3RL-SFS  | 3-key Rolls — Same Finger Skipgram  | "ted", "ill"        |
| RED      | Redirections total                  |                     |
| RED-WEAK | Redirections — Weak                 | "was", "ese"        |
| RED-SFS  | Redirections — Same Finger Skipgram | "you", "ter"        |
| RED-OTH  | Redirections — Other                | "ion", "ate", "ere" |
| IN:OUT   | Inward:Outward ratio                |          |
| FBL      | Finger Balance                      |          |
| POH      | Pinky Off Home                      |          |

### Metric details

Keycraft aims to follow the Keyboard Layouts Doc (KLD). 

#### Bigrams
- SFB - percentage of bigrams typed using the same finger (excluding identical-key repeats)
- LSB - percentage of bigrams that map to pre-defined lateral-stretch finger pairs
- FSB - percentage of bigrams forming pre-defined full-scissor patterns (>1.5U vertical separation)
- HSB - percentage of bigrams forming pre-defined half-scissor patterns (<=1.5U vertical separation)

#### Skipgrams
- SFS - percentage of skipgrams typed using the same finger (excluding identical-key skips)
- LSS - percentage of skipgrams that map to lateral-stretch pairs
- FSS - percentage of skipgrams forming full-scissor patterns
- HSS - percentage of skipgrams forming half-scissor patterns

#### Trigrams

Alternations - First key on one hand, the second key on the other, the last key on the first hand again
- ALT - total percentage of hand alternations (ALT-OTH + ALT-SFS)
- ALT-SFS - portion of cross-hand trigram alternations that are same‑finger alternations (excluding identical-key skips)
- ALT-OTH - portion of cross-hand trigram alternations not classified as SFS (normal alts)

Two-rolls - Two keys on one hand (rolling in- or outward), and one on the other (or vv)
- 2RL - total percentage for two-key in- and out-rolls (2RL-IN + 2RL-OUT)
- 2RL-IN - two-key roll trigrams classified as inward rolls
- 2RL-OUT - two-key roll trigrams classified as outward rolls
- 2RL-SFB - two-key roll trigrams where both keys use the same finger (any key). Argueably, identical-key repeats are not uncomfortable and should be separated out.

Three-rolls - All three keys are typed on one hand
- 3RL - total percentage for three-key in- and out-rolls (3RL-IN + 3RL-OUT)
- 3RL-IN - three-key roll trigrams classified as inward sequences
- 3RL-OUT - three-key roll trigrams classified as outward sequences
- 3RL-SFS - three-key roll trigrams where the first and last keys use the same finger (any key). Argueably, identical-key skips are not uncomfortable and should be separated out.

Redirections - All three keys on one hand
- RED - total percentage of redirections
- RED-WEAK - all redirections on one hand with no index involvement (weaker/bad redirections)
- RED-SFS - redirections on one hand that are same finger skipgrams (excluding identical-key skips)
- RED-OTH - other (normal) redirections on one hand

#### Other Metrics
- IN:OUT  - ratio of inward rolls to outward rolls computed as (2RL-IN + 3RL-IN) / (2RL-OUT + 3RL-OUT)
- FBL - cumulative absolute deviation (percentage points) from the ideal finger-load distribution
- POH - percentage of unigram frequency typed with a pinky while that pinky is off its home row

##### Ideal Finger-load distribution

|               | Left-Pinky | Left-Ring | Left-Middle | Left-Index | Right-Index | Right-Middle | Right-Ring | Right-Pinky |
|--------------:|:----------:|:---------:|:-----------:|:----------:|:-----------:|:------------:|:----------:|:-----------:|
| Ideal load (%)| 8.0        | 11.0      | 16.0        | 15.0       | 15.0        | 16.0         | 11.0       | 8.0         |

#### Hand balance metrics

This does not include space because n-grams with spaces are discarded from the corpus.

- H0, H1 (Hand usage) - percentage of total keystrokes by each hand (H0 = left, H1 = right). 
- F0–F9 (Finger usage) - percentage of total keystrokes by each finger (F0 = left pinky … F4 = left thumb, F5 = right thumb … F9 = right pinky).
- C0–C11 (Column usage) - percentage of total keystrokes per physical column on the layout.
- R0–R3 (Row usage) - percentage of total keystrokes per physical row on the layout.

## Usage

### Getting help

Use the `help` command to get help for the tool or a specific command. Most commands and flags have shortened versions to avoid typing too much.

```bash
# Get help on available commands
keycraft help

# Get help on a specific commands
keycraft help optimise

# Shortened version of the above
keycraft h o
```

Example output:

```bash
NAME:
   keycraft - A CLI tool for crafting better keyboard layouts

USAGE:
   keycraft [global options] command [command options]

COMMANDS:
   view, v        Analyse and display one or more keyboard layouts
   analyse, a     Analyse one or more keyboard layouts in detail
   rank, r        Rank keyboard layouts and optionally view deltas
   optimise, o    Optimise a keyboard layout
   experiment, x  Run experiments (for the developer)
   help, h        Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help
```

### Viewing one or more layouts

Use the `view` command and specify the layout(s) you want to view.

```bash
# View 1 or more layouts side-by-side
keycraft view focal.klf gallium-v2.klf

# View layouts with metrics based on another corpus
keycraft view -c monkeyracer.txt focal.klf gallium-v2.klf
```

- The layouts must be located in `./data/layouts`. To view your own layout, add the `.klf` file for your layout there.
- The corpus that is used to generate the stats is `./data/corpus/default.txt`. At the moment this is Shai's Cleaned iweb (90m words), available from:
  https://colemak.com/pub/corpus/iweb-corpus-samples-cleaned.txt.xz
- The first time a corpus is used (or after a corpus has changed), a cache is generated that will make loading it faster next time.

### Analysing and comparing one or more layouts

Use the `analyse` command and specify the layout(s) you want to analyse.

```bash
# Analyse multiple layouts with detailed tables for each metric
keycraft analyse focal.klf sturdy.klf gallium-v2.klf
```

- The trigram totals listed in the tables will be higher than the stats in the overview. This is because the tables show more than what is included in the overview. That's why they're called __detailed__ tables!

```
╭     ┬                                                      ┬                                                       ╮
                                FOCAL                                                  STURDY                         
├     ┼                                                      ┼                                                       ┤
 Board          ╭───┬───┬───╮          ╭───┬───┬───╮          ╭───┬───┬───┬───┬───┬───╮  ╭───┬───┬───┬───┬───┬───╮    
        ╭───┬───┤ l │ h │ g ├───╮  ╭───┤ f │ o │ u ├───┬───╮  │   │ v │ m │ l │ c │ p │  │ x │ f │ o │ u │ j │   │    
        │   │ v ├───┼───┼───┤ k │  │ q ├───┼───┼───┤ j │   │  ╰┬──┴┬──┴┬──┴┬──┴┬──┴┬──┴╮ ╰┬──┴┬──┴┬──┴┬──┴┬──┴┬──┴╮   
        ├───┼───┤ r │ n │ t ├───┤  ├───┤ c │ a │ e ├───┼───┤   │   │ s │ t │ r │ d │ y │  │ . │ n │ a │ e │ i │   │   
        │   │ s ├───┼───┼───┤ b │  │ y ├───┼───┼───┤ i │ / │   ╰─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─╮╰─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─┬─┴─╮ 
        ├───┼───┤ x │ m │ d ├───┤  ├───┤ w │ , │ ; ├───┼───┤     │   │ z │ k │ q │ g │ w │  │ b │ h │ ' │ ; │ , │   │ 
        │   │ z ├───┼───┼───┤ p │  │ ' ├───┼───┼───┤ . │   │     ╰───┴───┴───┼───┼───┼───┤  ├───┼───┼───┼───┴───┴───╯ 
        ╰───┴───╯   │   │   ├───┤  ├───┤   │   │   ╰───┴───╯                 │   │   │   │  │ _ │   │   │             
                    ╰───┴───┤   │  │ _ ├───┴───╯                             ╰───┴───┴───╯  ╰───┴───┴───╯             
                            ╰───╯  ╰───╯                                                                              
...
 Stats     ╭───────────┬───────────┬──────────┬──────────╮        ╭───────────┬───────────┬──────────┬──────────╮     
           │SFB:  0.76%│LSB:  1.46%│FSB: 0.07%│HSB: 3.54%│        │SFB:  0.90%│LSB:  1.25%│FSB: 0.28%│HSB: 2.88%│     
           ├───────────┼───────────┼──────────┼──────────┤        ├───────────┼───────────┼──────────┼──────────┤     
           │SFS:  6.48%│LSS:  1.51%│FSS: 0.74%│HSS: 5.45%│        │SFS:  5.54%│LSS:  2.08%│FSS: 0.23%│HSS: 4.60%│     
           ├───────────┼───────────┼──────────┼──────────┤        ├───────────┼───────────┼──────────┼──────────┤     
           │ALT: 39.10%│2RL: 44.51%│3RL: 1.39%│RED: 4.35%│        │ALT: 35.03%│2RL: 46.87%│3RL: 2.04%│RED: 5.14%│     
           ├───────────┼───────────┼──────────┼──────────┤        ├───────────┼───────────┼──────────┼──────────┤     
           │I:O:   1.00│FBL: 17.61%│POH: 2.44%│WEAK 0.23%│        │I:O:   0.94│FBL: 13.11%│POH: 2.35%│WEAK 0.32%│     
           ╰───────────┴───────────┴──────────┴──────────╯        ╰───────────┴───────────┴──────────┴──────────╯     
├     ┼                                                      ┼                                                       ┤
 SFB          ╭─────┬─────┬──────┬───────────┬───────╮                ╭────┬─────┬──────┬───────────┬───────╮         
              │     │ SFB │ DIST │     COUNT │     % │                │    │ SFB │ DIST │     COUNT │     % │         
              ├─────┼─────┼──────┼───────────┼───────┤                ├────┼─────┼──────┼───────────┼───────┤         
              │   1 │ ue  │ 1.00 │   440,608 │ 0.13% │                │  1 │ ue  │ 1.03 │   440,608 │ 0.13% │         
              │   2 │ pt  │ 1.49 │   267,309 │ 0.08% │                │  2 │ n.  │ 1.00 │   321,343 │ 0.09% │         
              │   3 │ rl  │ 1.00 │   261,012 │ 0.08% │                │  3 │ rl  │ 1.03 │   261,012 │ 0.08% │         
              │   4 │ oa  │ 1.00 │   254,172 │ 0.07% │                │  4 │ oa  │ 1.03 │   254,172 │ 0.07% │         
              │   5 │ cy  │ 1.00 │   110,370 │ 0.03% │                │  5 │ nf  │ 1.03 │   206,315 │ 0.06% │         
              │   6 │ o,  │ 2.00 │    90,677 │ 0.03% │                │  6 │ dy  │ 1.00 │   168,246 │ 0.05% │         
              │   7 │ dg  │ 2.00 │    89,160 │ 0.03% │                │  7 │ cy  │ 1.60 │   110,370 │ 0.03% │         
              │   8 │ a,  │ 1.00 │    84,158 │ 0.02% │                │  8 │ tm  │ 1.03 │    95,503 │ 0.03% │         
              │   9 │ hn  │ 1.00 │    78,587 │ 0.02% │                │  9 │ h.  │ 1.80 │    90,645 │ 0.03% │         
              │  10 │ eu  │ 1.00 │    68,749 │ 0.02% │                │ 10 │ dg  │ 1.12 │    89,160 │ 0.03% │         
              ├─────┼─────┼──────┼───────────┼───────┤                ├────┼─────┼──────┼───────────┼───────┤         
              │     │     │      │ 2,611,777 │ 0.76% │                │    │     │      │ 3,101,224 │ 0.90% │         
              ╰─────┴─────┴──────┴───────────┴───────╯                ╰────┴─────┴──────┴───────────┴───────╯         
...
 2RL    ╭───────┬─────┬──────┬─────────────┬────────┬──────╮    ╭───────┬─────┬──────┬─────────────┬────────┬──────╮  
        │       │ 2RL │ DIST │       COUNT │      % │ KIND │    │       │ 2RL │ DIST │       COUNT │      % │ KIND │  
        ├───────┼─────┼──────┼─────────────┼────────┼──────┤    ├───────┼─────┼──────┼─────────────┼────────┼──────┤  
        │     1 │ the │ 0.00 │   6,802,477 │  2.63% │ OUT  │    │     1 │ the │ 0.00 │   6,802,477 │  2.63% │ OUT  │  
        │     2 │ ing │ 0.00 │   3,233,466 │  1.25% │ IN   │    │     2 │ ing │ 0.00 │   3,233,466 │  1.25% │ IN   │  
        │     3 │ and │ 0.00 │   3,084,446 │  1.19% │ IN   │    │     3 │ and │ 0.00 │   3,084,446 │  1.19% │ IN   │  
        │     4 │ ion │ 0.00 │   1,720,124 │  0.67% │ IN   │    │     4 │ ent │ 0.00 │   1,528,636 │  0.59% │ IN   │  
        │     5 │ ent │ 0.00 │   1,528,636 │  0.59% │ IN   │    │     5 │ for │ 0.00 │   1,468,675 │  0.57% │ OUT  │  
        │     6 │ for │ 0.00 │   1,468,675 │  0.57% │ OUT  │    │     6 │ you │ 0.00 │   1,424,616 │  0.55% │ OUT  │  
        │     7 │ tio │ 0.00 │   1,380,462 │  0.53% │ IN   │    │     7 │ tio │ 0.00 │   1,380,462 │  0.53% │ IN   │  
        │     8 │ tha │ 0.00 │   1,189,955 │  0.46% │ OUT  │    │     8 │ hat │ 0.00 │   1,232,975 │  0.48% │ OUT  │  
        │     9 │ all │ 0.00 │   1,015,889 │  0.39% │ SF   │    │     9 │ tha │ 0.00 │   1,189,955 │  0.46% │ OUT  │  
        │    10 │ thi │ 0.00 │     936,782 │  0.36% │ OUT  │    │    10 │ her │ 0.00 │   1,158,203 │  0.45% │ OUT  │  
        ├───────┼─────┼──────┼─────────────┼────────┼──────┤    ├───────┼─────┼──────┼─────────────┼────────┼──────┤  
        │       │     │      │ 132,151,717 │ 51.15% │      │    │       │     │      │ 136,696,023 │ 52.91% │      │  
        ╰───────┴─────┴──────┴─────────────┴────────┴──────╯    ╰───────┴─────┴──────┴─────────────┴────────┴──────╯  
```

### Ranking layouts

Use the `rank` command to rank and compare a large number of layouts. Layouts are ranked by their overall score which depends on the weights you assign to each of the metrics, as well as the corpus you use. The weights that are applied are shown in the table's header.

```bash
# Rank all layouts in ./data/layouts
keycraft rank

# Rank all, showing the differences between each pair of rows, to easily compare layouts
keycraft rank --deltas rows

# Rank all, adding a "median" layout (ranked #0), showing the differences between this median layout and the other layouts
keycraft rank --deltas median

# Rank all, showing the differences between `canary.klf` and the other layouts
keycraft rank --deltas canary.klf

# Rank specific layouts only, comparing them against `canary.klf`
keycraft rank --deltas canary.klf colemak.klf colemak-qix.klf colemak-dh.klf

# Rank all, showing more columns with more metrics
keycraft rank --metrics extended

# Rank all, overriding the weight of the SFB metric
# Specifying a high value like below will effectively rank layouts based on SFBs only. Note the minus (-) sign!
keycraft rank --weights sfb=-1000
```

- Better layouts appear at the top of the list. `qwerty.klf` appears at the bottom of the list!
- The median layout is determined by taking the median of all layouts for each metric, normalising all metrics, and calculating the median layout's score by applying weights.
- Default weights are specified in the file `./data/weighs/default.txt`. You can either specify a different weights file using the `--weights-file` flag, or override specific weights using the `--weights` flag.
- The metrics shown in the table do not affect the calculation of each layout's score. The basic table for example shows a weight of 0 for ALT, which makes you think Alts do not affect the scores. But actually if you show extended metrics, you will see a weight is applied to ALT-SFS.

### Optimising a layout

Use the `optimise` command and specify the layout you want to optimise.

```bash
# Optimise a layout with an adjusted number of generations, but default weights and default accept-worse function ("drop-slow")
# "drop-slow" allows the optimisation engine to make radical jumps to very different layouts
keycraft optimise -g 500 qwerty.klf

# Optimise an already very good layout with the "never" accept-worse function and the some keys pinned
# "never" prevents the optimisation engine from making unpredictable jumps to worse layouts
# Pinning keys prevents them from being moved around, which could ruin the essence of a layout
keycraft optimise -g 100 -aw never --pins srntaeiou focal.klf

# Optimise a layout, strongly aiming for good finger balance, but potentially ruining other properties
keycraft optimise --weights FBL=-100 -aw never -g 100 canary.klf

# Optimise a small number of keys using the --free flag
# Optimising special characters should be used in combination with a more specific corpus
keycraft optimise -aw never -g 50 --free "';,.-/" graphite.klf
```

The result of fine-tuning Focal is shown below. Disregarding potentially negative changes, the lateral streches problem in this layout has been solved! Further adjustment of the weights may help against the negative effects.

```
╭     ┬                                                      ┬                                                      ╮
                                FOCAL                                                FOCAL-OPT                       
├     ┼                                                      ┼                                                      ┤
 Board          ╭───┬───┬───╮          ╭───┬───┬───╮                   ╭───┬───┬───╮          ╭───┬───┬───╮          
        ╭───┬───┤ l │ h │ g ├───╮  ╭───┤ f │ o │ u ├───┬───╮   ╭───┬───┤ l │ h │ g ├───╮  ╭───┤ w │ o │ u ├───┬───╮  
        │   │ v ├───┼───┼───┤ k │  │ q ├───┼───┼───┤ j │   │   │   │ v ├───┼───┼───┤ k │  │ q ├───┼───┼───┤ j │   │  
        ├───┼───┤ r │ n │ t ├───┤  ├───┤ c │ a │ e ├───┼───┤   ├───┼───┤ r │ n │ t ├───┤  ├───┤ c │ a │ e ├───┼───┤  
        │   │ s ├───┼───┼───┤ b │  │ y ├───┼───┼───┤ i │ / │   │   │ s ├───┼───┼───┤ f │  │ . ├───┼───┼───┤ i │ y │  
        ├───┼───┤ x │ m │ d ├───┤  ├───┤ w │ , │ ; ├───┼───┤   ├───┼───┤ x │ m │ d ├───┤  ├───┤ p │ ' │ ; ├───┼───┤  
        │   │ z ├───┼───┼───┤ p │  │ ' ├───┼───┼───┤ . │   │   │   │ z ├───┼───┼───┤ b │  │ , ├───┼───┼───┤ / │   │  
        ╰───┴───╯   │   │   ├───┤  ├───┤   │   │   ╰───┴───╯   ╰───┴───╯   │   │   ├───┤  ├───┤   │   │   ╰───┴───╯  
                    ╰───┴───┤   │  │ _ ├───┴───╯                           ╰───┴───┤   │  │ _ ├───┴───╯              
                            ╰───╯  ╰───╯                                           ╰───╯  ╰───╯                      
...
╭─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────╮
│                                                                                Layout Ranking                                                                               │
├───┬───────────────┬───────┬────────┬────────┬────────┬────────┬────────┬────────┬────────┬────────┬────────┬────────┬────────┬────────┬──────────┬────────┬────────┬────────┤
│ # │ NAME          │ SCORE │    SFB │    LSB │    FSB │    HSB │    SFS │    LSS │    FSS │    HSS │    ALT │    2RL │    3RL │    RED │ RED-WEAK │ IN:OUT │    FBL │    POH │
│   │ WEIGHT        │       │  -6.00 │  -4.00 │  -3.00 │  -3.00 │  -1.50 │  -1.00 │  -0.75 │  -0.75 │   0.00 │   0.00 │   0.00 │  -3.00 │     0.00 │   0.00 │  -3.00 │  -4.00 │
├───┼───────────────┼───────┼────────┼────────┼────────┼────────┼────────┼────────┼────────┼────────┼────────┼────────┼────────┼────────┼──────────┼────────┼────────┼────────┤
│ 1 │ focal-opt.klf │ +9.06 │  0.77% │  0.72% │  0.08% │  3.58% │  6.44% │  2.14% │  0.65% │  5.06% │ 40.15% │ 43.32% │  1.05% │  4.80% │    0.89% │   1.05 │ 18.09% │  1.37% │
│   │               │       │ -0.01% │ +0.74% │ -0.00% │ -0.04% │ +0.04% │ -0.63% │ +0.09% │ +0.39% │ -1.05% │ +1.19% │ +0.34% │ -0.45% │   -0.66% │  -0.05 │ -0.48% │ +1.07% │
│ 2 │ focal.klf     │ +4.70 │  0.76% │  1.46% │  0.07% │  3.54% │  6.48% │  1.51% │  0.74% │  5.45% │ 39.10% │ 44.51% │  1.39% │  4.35% │    0.23% │   1.00 │ 17.61% │  2.44% │
╰───┴───────────────┴───────┴────────┴────────┴────────┴────────┴────────┴────────┴────────┴────────┴────────┴────────┴────────┴────────┴──────────┴────────┴────────┴────────╯
```

## Configuration

### Specifying and choosing a suitable corpus (for all commands)

More information will be provided.

### Specifying weights (for ranking and optimising)

- Describe config locations, file format (YAML/JSON), and common options.
	layoutDir  = "data/layouts/"
	corpusDir  = "data/corpus/"
	weightsDir = "data/weights/"
	pinsDir    = "data/pins/"

## Contributing
- Questions, suggestions, and feedback are super welcome! Just open a New Issue and I'll get back to you as soon as I can.
- I probably cannot take PRs until I feel a solid base implementation is in place.

## License
BSD-3-Clause license. See LICENSE file for details.

## Contact
- Author: Barend Scholtus <barend.scholtus@gmail.com>
- Issue tracker: https://github.com/rbscholtus/keycraft/issues
