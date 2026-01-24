# wt

`wt` is a small Git worktree helper written in Go. It focuses on fast
worktree creation, simple hooks, and shell integration for `cd`.

## Install

Homebrew (tap):

```
brew install hiroki23/tap/wt
```

## Setup (zsh)

Add this to your `.zshrc`:

```
eval "$(wt hook zsh)"
```

This enables `wt cd` and `wt co` to change directories by running
the `wt` binary and `cd`-ing to its output.

Prompt integration is opt-in:

```
eval "$(wt hook zsh --prompt)"
```

## Quick Start

```
wt init
wt add feat/one
wt co feat/one
wt cd -
```

## Commands

- `wt init [-g]`  
  Create `.wt.yaml` (local or global).

- `wt add <branch>`  
  Create a worktree for the branch.

- `wt co <branch>`  
  Create the worktree (if needed) and output its path.
  With the zsh hook, this also `cd`s to the path.

- `wt cd [branch|-]`  
  Output the worktree path for the branch. No args goes to the main
  worktree. `-` outputs `-` for `cd -`.

- `wt rm <branch>`  
  Remove the worktree and delete the branch.
  The default and current branches are protected.

- `wt rm --all [-f]`  
  Remove all non-main worktrees (and their branches).  
  This must be run from the main worktree. Use `-f` to skip confirmation.

- `wt list`  
  List worktrees.

- `wt prune`  
  Run `git worktree prune`.

## Configuration

Local config: `<git root>/.wt.yaml`  
Global config: `~/.wt.yaml`

Local config has priority over global.

Example:

```yaml
version: 1
base_dir: "../worktrees/{gitroot}"

hooks:
  post_create:
    copy:
      - ".env"
      - ".env.local"
    symlink:
      - "storage"
    run:
      - "bundle install"
```

### `base_dir`

Default: `../worktrees/{gitroot}`

`{gitroot}` is replaced with the repository directory name. Paths are
resolved from the main worktree.

### Hooks

Hooks are applied after a worktree is created.

- `copy`: copy files or directories from the main worktree
- `symlink`: create symlinks from the main worktree
- `run`: run commands in the new worktree

## Prompt

Enable with `wt hook zsh --prompt`.

Customize with `WT_PROMPT_FORMAT`:

- `%w` base path (before branch)
- `%b` branch name
- `%r` relative path inside worktree
- `%s` git status marker (`*` when dirty)
- `%%` literal `%`

Default format:

```
%F{cyan}%w@%F{yellow}%b%F{cyan}%r%f %(!.#.$) 
```

If you need deeper customization, keep your own prompt functions and
skip `--prompt`.

## Notes

`wt` only changes directories when the shell hook is enabled. Without
the hook, `wt cd` and `wt co` just print paths.
