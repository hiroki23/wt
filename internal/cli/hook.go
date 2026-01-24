package cli

import (
	"fmt"
	"io"
)

func runHook(args []string, out io.Writer) int {
	if len(args) < 1 || len(args) > 2 {
		return printCommandUsage(out, "hook")
	}
	if args[0] != "zsh" {
		fmt.Fprintln(out, "wt hook: unsupported shell")
		return 2
	}

	withPrompt := false
	if len(args) == 2 {
		if args[1] != "--prompt" {
			return printCommandUsage(out, "hook")
		}
		withPrompt = true
	}

	fmt.Fprintln(out, hookZshScript(withPrompt))
	return 0
}

func hookZshScript(withPrompt bool) string {
	script := `# wt hook for zsh
_wt_run() {
  command wt "$@"
}

wt() {
  if [ "$1" = "cd" ] || [ "$1" = "co" ]; then
    local out
    out="$(_wt_run "$@")"
    if [ "$?" -ne 0 ]; then
      echo "$out"
      return 1
    fi
    if [ -n "$out" ]; then
      builtin cd "$out" || return $?
    fi
    return 0
  fi
  _wt_run "$@"
}
`
	if !withPrompt {
		return script
	}

	script += `

# wt prompt (opt-in via --prompt)
setopt PROMPT_SUBST
setopt prompt_percent

_WT_PROMPT_DEFAULT='%F{cyan}%w@%F{yellow}%b%F{cyan}%r%f %(!.#.$) '

_wt_prompt_escape() {
  local value="$1"
  value="${value//\%/%%}"
  print -r -- "$value"
}

_wt_prompt_status() {
  if git diff --quiet --ignore-submodules -- 2>/dev/null && \
     git diff --quiet --ignore-submodules --cached -- 2>/dev/null && \
     [ -z "$(git ls-files --others --exclude-standard 2>/dev/null)" ]; then
    print -r -- ""
  else
    print -r -- "*"
  fi
}

_wt_prompt_format() {
  local format="$1"
  local base="$2"
  local branch="$3"
  local rel="$4"
  local wt_status="$5"

  local esc_base="$(_wt_prompt_escape "$base")"
  local esc_branch="$(_wt_prompt_escape "$branch")"
  local esc_rel="$(_wt_prompt_escape "$rel")"
  local esc_status="$(_wt_prompt_escape "$wt_status")"

  local out="$format"
  out="${out//\%\%/__WT_PCT__}"
  out="${out//\%w/$esc_base}"
  out="${out//\%b/$esc_branch}"
  out="${out//\%r/$esc_rel}"
  out="${out//\%s/$esc_status}"
  out="${out//__WT_PCT__/%%}"
  print -r -- "$out"
}

_wt_update_prompt() {
  if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    PROMPT="$_WT_PROMPT_BASE"
    return
  fi

  local root
  root="$(git rev-parse --show-toplevel 2>/dev/null)" || {
    PROMPT="$_WT_PROMPT_BASE"
    return
  }

  if [ ! -f "$root/.git" ]; then
    PROMPT="$_WT_PROMPT_BASE"
    return
  fi

  local branch
  branch="$(git branch --show-current 2>/dev/null)"
  if [ -z "$branch" ]; then
    branch="$(git rev-parse --short HEAD 2>/dev/null)"
  fi

  local base
  base="${root%/$branch}"
  if [ "$base" = "$root" ]; then
    base="$(dirname "$root")"
  fi
  local base_display="${base/#$HOME/~}"
  local rel="${PWD#$root}"

  local format="${WT_PROMPT_FORMAT:-$_WT_PROMPT_DEFAULT}"
  local wt_status=""
  if [[ "$format" == *"%s"* ]]; then
    wt_status="$(_wt_prompt_status)"
  fi

  PROMPT="$(_wt_prompt_format "$format" "$base_display" "$branch" "$rel" "$wt_status")"
}

typeset -g _WT_PROMPT_BASE
if [ -z "$_WT_PROMPT_BASE" ]; then
  _WT_PROMPT_BASE="$PROMPT"
fi

precmd_functions+=(_wt_update_prompt)
`
	return script
}
