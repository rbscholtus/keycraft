# Shell Completion Setup Guide

Keycraft supports dynamic shell completion for bash, zsh, fish, and PowerShell, making it easier to use commands and flags.

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
keycraft rank --<TAB>       # Shows all flags for the rank command
keycraft --<TAB>            # Shows global flags
```

### Layout File Completion

All commands that work with keyboard layouts support tab completion for `.klf` files:

```bash
keycraft view <TAB>         # Complete layout file names
keycraft analyse apt<TAB>   # Complete to "apt-v3" or "apt-v4"
keycraft rank qwer<TAB>     # Complete to "qwerty"
keycraft flip colemak<TAB>  # Complete to "colemak-dh"
keycraft optimise <TAB>     # Complete layout names
```

**Supported commands**: `view`, `analyse`, `rank`, `flip`, `optimise`

The completion automatically:
- Lists all `.klf` files from `data/layouts/` directory
- Shows filenames without the `.klf` extension for cleaner typing
- Works with partial matches (type the first few letters and press TAB)

### Corpus File Completion

All commands that accept the `--corpus` flag support tab completion for corpus files:

```bash
keycraft corpus --corpus <TAB>     # Complete corpus file names
keycraft view --corpus <TAB>       # Complete corpus file names
keycraft analyse -c def<TAB>       # Complete to "default.txt"
keycraft rank --corpus <TAB>       # Complete corpus file names
keycraft optimise -c <TAB>         # Complete corpus file names
```

**Supported commands**: `corpus`, `view`, `analyse`, `rank`, `optimise`

The completion automatically:
- Lists all corpus files from `data/corpus/` directory
- Handles both source files (`.txt`) and cached files (`.txt.json`)
- Deduplicates files (shows "english.txt" even if both "english.txt" and "english.txt.json" exist)
- Filters out hidden files (files starting with `.`)

## Features

With shell completion enabled, you can:

- **Tab-complete commands**: Type the first few letters of a command and press TAB
- **Tab-complete flags**: Type `--` and press TAB to see all available flags
- **Tab-complete subcommands**: Navigate through command hierarchies with TAB
- **Get command suggestions**: Mistyped commands will show helpful "Did you mean?" suggestions

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

### Flag completion behavior

**Zsh users:** When typing flags, the completion workflow is:
- Type `keycraft corpus -<TAB>` → completes to `--` (common prefix)
- Type `keycraft corpus --<TAB>` → shell shows flag menu (may need to type more)
- Type `keycraft corpus --co<TAB>` → completes to `--corpus` or `--coverage`

**Bash users:** Flag completion filters automatically:
- Type `keycraft corpus --co<TAB>` → immediately shows `--corpus` and `--coverage`

This difference is by design - zsh's completion system requires partial matches, while bash's `compgen` automatically filters based on what you've typed.

### Old version still active

If you see "No help topic for 'completion'", you may have an old version of keycraft in your PATH. Update it with:

```bash
go install ./cmd/keycraft
```

Then verify the version:

```bash
keycraft --version
```

## Additional Features

Keycraft also provides:

- **Command suggestions**: When you mistype a command, keycraft suggests the correct one
  ```bash
  $ keycraft optimze
  Command 'optimze' is not a thing. Did you mean: optimise?
  ```

- **Helpful error messages**: Clear guidance when commands aren't found
  ```bash
  $ keycraft xyz
  Command 'xyz' is not a thing. Run 'keycraft help' to see available commands.
  ```
