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
         │ALT: 26.11%│.NML: 20.65%│.SFS:  5.45%│           │   
         ├───────────┼────────────┼────────────┼───────────┤   
         │2RL: 46.42%│.IN:  19.85%│.OUT: 16.97%│.SFB: 9.60%│   
         ├───────────┼────────────┼────────────┼───────────┤   
         │3RL: 11.82%│.IN:   1.29%│.OUT:  1.45%│.SFB: 9.08%│   
         ├───────────┼────────────┼────────────┼───────────┤   
         │RED: 13.04%│.NML:  7.38%│.WEAK: 1.41%│.SFS: 4.25%│   
         ├───────────┼────────────┼────────────┼───────────┤   
         │I:O:   1.15│FLW:  60.21%│            │           │   
         ├───────────┼────────────┼────────────┼───────────┤   
         │RBL:  81.52│FBL:  24.27%│POH:   4.80%│           │   
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
│   ├── layouts
│   │   ├── ...
│   │   └── ...
│   ├── pins
│   │   ├── focal.pin
│   │   └── qwerty.pin
│   └── weights
│       └── default.txt
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
# View corpus statistics for the default corpus (shai)
keycraft c

# View a built-in layout
keycraft v qwerty

# Analyse and compare multiple layouts
keycraft a qwerty colemak dvorak

# Rank built-in layouts with custom weights (making SFB the dominant metric)
# Note many metrics are considered negative are written with a -
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
- Supports 56 layout metrics (including row, column, and finger balance)
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
| ALT-NML  | Alternation — Other                 | "and", "ent", "iti" |
| 2RL      | 2-key Rolls total                   |                     |
| 2RL-IN   | 2-key Rolls — Inward                | "ing", "hat"        |
| 2RL-OUT  | 2-key Rolls — Outward               | "tio", "thi"        |
| 2RL-SFB  | 2-key Rolls — Same Finger Bigram    | "nce", "all"        |
| 3RL      | 3-key Rolls total                   |                     |
| 3RL-IN   | 3-key Rolls — Inward                | "act", "lin"        |
| 3RL-OUT  | 3-key Rolls — Outward               | "rea", "tes"        |
| 3RL-SFB  | 3-key Rolls — Same Finger Bigram    | "ted", "ill"        |
| RED      | Redirections total                  |                     |
| RED-WEAK | Redirections — Weak                 | "was", "ese"        |
| RED-SFS  | Redirections — Same Finger Skipgram | "you", "ter"        |
| RED-NML  | Redirections — Other                | "ion", "ate", "ere" |
| IN:OUT   | Inward:Outward rolls ratio          |          |
| FLW      | Flowiness                           |          |
| H0,H1    | Left-hand, Right-hand usage         |          |
| RBL      | Row Balance                         |          |
| FBL      | Finger Balance                      |          |
| POH      | Pinky Off Home (Weighted)           |          |
| Not shown | Row, column, and finger balance    |          |



### Metric details

Keycraft aims to follow the Keyboard Layouts Doc (KLD) as much as possible.
Some metrics are unique to Keycraft though: RBL, FBL, POH, FLW

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

- ALT - total percentage of hand alternations (ALT-NML + ALT-SFS)
- ALT-SFS - portion of cross-hand trigram alternations that are same‑finger alternations (excluding identical-key skips)
- ALT-NML - portion of cross-hand trigram alternations not classified as SFS (normal alts)

Two-rolls - Two keys on one hand, and one on the other (or vv)

- 2RL - total percentage of two-key roll trigrams (2RL-IN + 2RL-OUT + 2RL-SFB)
- 2RL-IN - two-key roll trigrams classified as inward rolls
- 2RL-OUT - two-key roll trigrams classified as outward rolls
- 2RL-SFB - two-key roll trigrams where both keys use the same finger (any key). Argueably, identical-key repeats are not uncomfortable and should be separated out.

Three-rolls - All three keys are typed on one hand

- 3RL - total percentage of three-key roll trigrams (3RL-IN + 3RL-OUT + 3RL-SFB)
- 3RL-IN - three-key roll trigrams classified as inward sequences
- 3RL-OUT - three-key roll trigrams classified as outward sequences
- 3RL-SFB - three-key roll trigrams where the first and last keys use the same finger (any key). Argueably, identical-key skips are not uncomfortable and should be separated out.

Redirections - All three keys on one hand

- RED - total percentage of redirections
- RED-WEAK - all redirections on one hand with no index involvement (weaker/bad redirections)
- RED-SFS - redirections on one hand that are same finger skipgrams (excluding identical-key skips)
- RED-NML - other (normal) redirections on one hand

#### Other Metrics

- IN:OUT - ratio of inward rolls to outward rolls computed as (2RL-IN + 3RL-IN) / (2RL-OUT + 3RL-OUT)
- FLW - total percentage of normal alternations and inward and outward sequences
- RBL - cumulative deviation (percentage points) from some reference row-load distribution. More detail needed.
- FBL - cumulative absolute deviation (percentage points) from the ideal finger-load distribution
- POH - percentage of unigram frequency typed with a pinky while off the home row key, weighted using position-specific weights that penalize some pinky positions more heavily than others

##### Reference Row-load distribution

More detail coming.

|               | Top row | Home row | Bottom row |
|--------------:|:-------:|:--------:|:----------:|
| Reference (%) | 18.5    | 73.0     | 8.5        |

##### Ideal Finger-load distribution

More detail coming.

|               | Left-Pinky | Left-Ring | Left-Middle | Left-Index | Right-Index | Right-Middle | Right-Ring | Right-Pinky |
|--------------:|:----------:|:---------:|:-----------:|:----------:|:-----------:|:------------:|:----------:|:-----------:|
| Ideal load (%)| 7.5        | 11.0      | 16.0        | 15.5       | 15.5        | 16.0         | 11.0       | 7.5         |

#### Hand balance metrics

Most hand balance metrics are based on characters on the first 3 rows only. That means keys on thumbclusters are not counted. Further, keys found on the shift layer (for example "?" and ":" on QWERTY) are not counted either.

- H0, H1 (Hand usage) - percentage of total keystrokes for each hand (H0 = left, H1 = right).
- F0–F9 (Finger usage) - percentage of total keystrokes for each finger (F0 = left pinky … F3 = left index, F6 = right index … F9 = right pinky).
- C0–C11 (Column usage) - percentage of total keystrokes per physical column on the layout.
- R0–R3 (Row usage) - percentage of total keystrokes per physical row on the layout, including the thumb row.

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
- Default weights are specified in the file `./data/weighs/default.txt`. You can either specify a different weights file using the `--weights-file` flag, or override specific weights using the `--weights` flag.

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
 weightsDir = "data/weights/"
 pinsDir    = "data/pins/"

## Contributing

- Questions, suggestions, and feedback are super welcome! Just open a New Issue and I'll get back to you as soon as I can.
- I probably cannot take PRs until I feel a solid base implementation is in place. Sorry about this!

## License

BSD-3-Clause license. See LICENSE file for details.

## Contact

- Author: Barend Scholtus <barend.scholtus@gmail.com>
- Issue tracker: <https://github.com/rbscholtus/keycraft/issues>
- Discord: @ironcollar
