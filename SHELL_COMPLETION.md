# Shell Completion Setup Guide

Keycraft supports dynamic shell completion for bash, zsh, fish, and PowerShell, making it easier to use commands and flags.

## Features

With shell completion enabled, you can:

- **Tab-complete commands**: Type the first few letters of a command and press TAB
- **Tab-complete flags**: Type `--` followed by at least one letter and press TAB to see matching flags
- **Get command suggestions**: Mistyped commands will show helpful "Did you mean?" suggestions

## Installation

First, ensure you have the latest version of keycraft installed:

```bash
go install ./cmd/keycraft
```

## Setup Instructions

### Zsh

Add this line to your `~/.zshrc`:

```bash
source <(keycraft completion zsh)
```

Or generate and save the completion script:

```bash
keycraft completion zsh > ~/.config/zsh/completions/_keycraft
```

Then reload your shell:

```bash
source ~/.zshrc
```

### Bash

Add this line to your `~/.bashrc` or `~/.bash_profile`:

```bash
source <(keycraft completion bash)
```

Then reload your shell:

```bash
source ~/.bashrc
```

### Fish

Generate the completion script for fish:

```bash
keycraft completion fish > ~/.config/fish/completions/keycraft.fish
```

The completion will be available in new fish sessions.

### PowerShell

Generate the completion script:

```powershell
keycraft completion powershell > keycraft.ps1
```

Then run the script to enable completion (you may need to adjust your execution policy).

## Testing

After setting up completion, try these commands:

```bash
keycraft <TAB>              # Shows all available commands
keycraft opt<TAB>           # Completes to "optimise"
keycraft view <TAB>         # Shows available .klf layout files
keycraft corpus --corpus <TAB>  # Shows available corpus files
keycraft rank --co<TAB>     # Completes to --corpus or --coverage
```

**Note**: Typing just `--<TAB>` (without any letters after `--`) will not show flag completions due to a framework limitation. The CLI framework treats `--` as an argument terminator. To see available flags, type at least one letter after `--` (e.g., `--c<TAB>`).

### Layout File Completion

All commands that work with keyboard layouts support tab completion for `.klf` files, for example:

```bash
keycraft view <TAB>         # Complete layout file names
keycraft analyse ar<TAB>    # Complete to "arensito" or "arts"
keycraft rank qwer<TAB>     # Complete to "qwerty"
```

The completion automatically:
- Lists all `.klf` files from `data/layouts/` directory
- Shows filenames without the `.klf` extension for cleaner typing
- Works with partial matches (type the first few letters and press TAB)

### Corpus File Completion

All commands that accept the `--corpus` flag support tab completion for corpus files, for example:

```bash
keycraft corpus --corpus <TAB>     # Complete corpus file names
keycraft view -c <TAB>             # Complete corpus file names
keycraft analyse -c def<TAB>       # Complete to "default.txt"
```

The completion automatically:
- Lists all corpus files from `data/corpus/` directory
- Handles both source files (`.txt`) and cached files (`.txt.json`)
- Deduplicates files (shows "english.txt" even if both "english.txt" and "english.txt.json" exist)
- Filters out hidden files (files starting with `.`)

## Troubleshooting

### Completion not working

1. **Verify keycraft is in your PATH:**
   ```bash
   which keycraft
   ```
   If not found, install it with `go install ./cmd/keycraft`

2. **Check if completion is installed (zsh only):**
   ```bash
   which _keycraft
   ```
   If you see "_keycraft not found", the completion isn't loaded. Install it:
   ```bash
   # Quick test (temporary)
   source <(keycraft completion zsh)

   # Or permanent install
   keycraft completion zsh > ~/.zsh/completions/_keycraft
   # Add to ~/.zshrc if not already there:
   fpath=(~/.zsh/completions $fpath)
   autoload -Uz compinit && compinit
   # Reload shell:
   exec zsh
   ```

3. **Test the completion output manually:**
   ```bash
   keycraft completion bash  # or zsh, fish, powershell
   ```
   This should output the completion script.

4. **Verify completion works at the command level:**
   ```bash
   keycraft corpus --co --generate-shell-completion
   ```
   Should output: `--corpus:...` and `--coverage:...`

5. **Reload your shell configuration:**
   ```bash
   source ~/.bashrc  # or ~/.zshrc
   ```
