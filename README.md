# wt

`wt` is a small Git worktree helper written in Go. It focuses on fast
worktree creation, simple hooks, and shell integration for `cd`.

## Install

Homebrew (tap):

```
brew install hiroki23/tap/wt
```

## Setup (shell integration)

Add one of these to your shell config:

zsh (`~/.zshrc`):

```
eval "$(wt hook zsh)"
```

bash (`~/.bashrc`):

```
eval "$(wt hook bash)"
```

fish (`~/.config/fish/config.fish`):

```
wt hook fish | source
```

This enables `wt cd` and `wt co` to move between worktrees.

Prompt integration is zsh-only and opt-in:

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
  Create the worktree (if needed) and move to it.  
  Alias: `wt checkout <branch>`.

- `wt cd [branch|-]`  
  Move to the worktree for the branch. No args goes to the main
  worktree (git root). `-` goes back.
  Shorthand: `wt <branch>`.

- `wt rm <branch>`  
  Remove the worktree and delete the branch.
  The default and current branches are protected.  
  Alias: `wt remove <branch>`.

- `wt rm --all [-f]`  
  Remove all non-main worktrees (and their branches).  
  This must be run from the main worktree (git root). Use `-f` to skip confirmation.

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

If the shell hook is not enabled, `wt cd` and `wt co` print the target
path instead of moving.
