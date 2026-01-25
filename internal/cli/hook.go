package cli

import (
	"fmt"
	"io"
)

func runHook(args []string, out io.Writer) int {
	if len(args) < 1 || len(args) > 2 {
		return printCommandUsage(out, "hook")
	}

	shell := args[0]
	withPrompt := false
	if len(args) == 2 {
		if args[1] != "--prompt" {
			return printCommandUsage(out, "hook")
		}
		withPrompt = true
	}

	switch shell {
	case "zsh":
		fmt.Fprintln(out, hookZshScript(withPrompt))
		return 0
	case "bash":
		if withPrompt {
			fmt.Fprintln(out, "wt hook: prompt is only supported for zsh")
			return 2
		}
		fmt.Fprintln(out, hookBashScript())
		return 0
	case "fish":
		if withPrompt {
			fmt.Fprintln(out, "wt hook: prompt is only supported for zsh")
			return 2
		}
		fmt.Fprintln(out, hookFishScript())
		return 0
	default:
		fmt.Fprintln(out, "wt hook: unsupported shell")
		return 2
	}
}

func hookZshScript(withPrompt bool) string {
	script := `# wt hook for zsh
_wt_run() {
  command wt "$@"
}

_wt_is_cmd() {
  case "$1" in
    init|hook|add|co|cd|rm|list|prune|help)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

_wt_cd() {
  local out
  out="$(_wt_run "$@")"
  local exit_code="$?"
  if [ "$exit_code" -ne 0 ]; then
    echo "$out"
    return "$exit_code"
  fi
  if [ -n "$out" ]; then
    builtin cd "$out" || return $?
  fi
  return 0
}

wt() {
  if [ "$1" = "cd" ] || [ "$1" = "co" ]; then
    _wt_cd "$@"
    return $?
  fi
  if [ "$#" -eq 1 ] && [ "${1#-}" = "$1" ] && ! _wt_is_cmd "$1"; then
    _wt_cd "$@"
    return $?
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

func hookBashScript() string {
	return `# wt hook for bash
_wt_run() {
  command wt "$@"
}

_wt_is_cmd() {
  case "$1" in
    init|hook|add|co|cd|rm|list|prune|help)
      return 0
      ;;
    *)
      return 1
      ;;
  esac
}

_wt_cd() {
  local out
  out="$(_wt_run "$@")"
  local exit_code="$?"
  if [ "$exit_code" -ne 0 ]; then
    echo "$out"
    return "$exit_code"
  fi
  if [ -n "$out" ]; then
    builtin cd "$out" || return $?
  fi
  return 0
}

wt() {
  if [ "$1" = "cd" ] || [ "$1" = "co" ]; then
    _wt_cd "$@"
    return $?
  fi
  if [ "$#" -eq 1 ] && [ "${1#-}" = "$1" ] && ! _wt_is_cmd "$1"; then
    _wt_cd "$@"
    return $?
  fi
  _wt_run "$@"
}
`
}

func hookFishScript() string {
	return `# wt hook for fish
function _wt_run
  command wt $argv
end

function _wt_is_cmd
  switch $argv[1]
    case init hook add co cd rm list prune help
      return 0
  end
  return 1
end

function _wt_cd
  set -l out (_wt_run $argv)
  set -l code $status
  if test $code -ne 0
    echo $out
    return $code
  end
  if test -n "$out"
    cd "$out"; or return $status
  end
  return 0
end

function wt
  if test (count $argv) -ge 1
    set -l cmd $argv[1]
    if test "$cmd" = "cd" -o "$cmd" = "co"
      _wt_cd $argv
      return $status
    end
    if test (count $argv) -eq 1
      if not string match -rq '^-.*' -- $argv[1]
        if not _wt_is_cmd $argv[1]
          _wt_cd $argv
          return $status
        end
      end
    end
  end
  _wt_run $argv
end
`
}
