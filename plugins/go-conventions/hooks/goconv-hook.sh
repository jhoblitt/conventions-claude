#!/usr/bin/env bash
# Build-on-first-use launcher for the hook binary in hooks/goconv-hook/.
#
#   goconv-hook.sh format | session-start        (the hook JSON on stdin)
#
# The launcher contract -- and the rule that a HOOK launcher is the one that
# fails open -- is AGENTS.md, Plugin tools. Every failure path here exits 0 in
# silence: a broken build must never brick the tool call or the session this
# wraps.
#
# The PostToolUse form runs on every Write and Edit in the session, so the work
# is ordered by what it costs. stdin is read once and grepped for the shape
# that could possibly produce output; only a payload that survives that reaches
# a build, and only then the binary, which re-parses the JSON properly. The
# grep never decides to ACT -- it only ever declines, and only on a payload it
# can read literally. Anything it cannot read that way falls through to the
# binary, so a value the grep gets wrong costs a build, never an answer.
#
# The binary's exit status is discarded and this script always exits 0. Exit 2
# from a PostToolUse hook is Claude Code's BLOCKING status: propagating it
# would turn a panic in the binary into a hook that blocks every Write and Edit
# in the session. That is also the difference from a PreToolUse gate, which
# exits non-zero on purpose.
#
# Correctness under parallel hooks comes from the per-PID temp plus atomic mv;
# flock, where present, only dedupes the work.
#
# The kill switch is checked here as well as in Go: the escape hatch for a
# broken hook cannot live only inside the artifact whose build may be what
# broke.
set -uo pipefail

[ "${GOCONV_HOOK:-on}" = "off" ] && exit 0

mode=${1:-}
case "$mode" in
format | session-start) ;;
*) exit 0 ;;
esac

stdin=$(cat)

case "$mode" in
format)
  printf '%s' "$stdin" |
    grep -qE '"file_path"[[:space:]]*:[[:space:]]*"[^"]*\.go"' || exit 0
  ;;
session-start)
  # The event's cwd is the tree in play; CLAUDE_PROJECT_DIR still names where
  # the session started, which is the wrong tree after entering a worktree.
  dir=$(printf '%s' "$stdin" |
    grep -oE '"cwd"[[:space:]]*:[[:space:]]*"[^"]*"' |
    head -n 1 | sed -e 's/.*"[[:space:]]*:[[:space:]]*"//' -e 's/"$//')
  [ -n "$dir" ] || dir="${CLAUDE_PROJECT_DIR:-}"
  case "$dir" in
  # No cwd to test, or a JSON-escaped one this cannot turn back into a path:
  # unescaping here would be the grep deciding. Hand it to the binary.
  "" | *\\*) ;;
  *) [ -f "$dir/go.mod" ] || exit 0 ;;
  esac
  ;;
esac

src="${CLAUDE_PLUGIN_ROOT:-}/hooks/goconv-hook"
[ -n "${CLAUDE_PLUGIN_ROOT:-}" ] && [ -d "$src" ] || exit 0

# The binary lives in the per-plugin data directory, never in
# CLAUDE_PLUGIN_ROOT: the cache is a per-version directory that moves on every
# update and can be reaped mid-session. Data outlives versions, and a fresh
# version clone carries fresh mtimes, so the staleness check below rebuilds
# after an update on its own.
#
# Which basename that directory must carry is the contract cited above; the
# model runs hooks from its own shell, so CLAUDE_PLUGIN_DATA may hold ANOTHER
# installed plugin's directory. The name is a literal here rather than parsed
# out of .claude-plugin/plugin.json -- sed would silently pick author.name if
# the keys were reordered.
plugin=go-conventions
data="${XDG_CACHE_HOME:-${HOME:-}/.cache}/conventions-claude/$plugin"
if [ -n "${CLAUDE_PLUGIN_DATA:-}" ]; then
  case "$(basename "$CLAUDE_PLUGIN_DATA")" in
  "$plugin" | "$plugin"-*) data="$CLAUDE_PLUGIN_DATA/hooks" ;;
  esac
fi
bin="$data/goconv-hook"

stale() {
  [ ! -x "$bin" ] && return 0
  # The parens matter: without them -o binds looser than the implicit -a and
  # every .go file matches regardless of mtime, so the hook rebuilds on every
  # single invocation.
  [ -n "$(find "$src" -maxdepth 1 \( -name '*.go' -o -name go.mod \) \
    -newer "$bin" -print -quit 2>/dev/null)" ]
}

# A failed build must not retry on every Write. The marker pins the source
# fingerprint it failed on: a source change re-arms immediately (the dev
# iterating on a fix), expiry re-arms after transient toolchain trouble.
# Success removes it.
build() {
  local marker="$bin.buildfail" fp
  # The hook has a 30s timeout; a build killed by it would otherwise leave
  # $bin.tmp.$$ in the data directory forever.
  trap 'rm -f "$bin.tmp.$$"' EXIT
  fp=$(cat "$src"/go.mod "$src"/*.go 2>/dev/null | cksum)
  if [ -f "$marker" ] && [ "$(cat "$marker" 2>/dev/null)" = "$fp" ] &&
    [ -z "$(find "$marker" -mmin +60 -print 2>/dev/null)" ]; then
    return 1
  fi
  if (cd "$src" && go build -o "$bin.tmp.$$" .) >/dev/null 2>&1 &&
    mv -f "$bin.tmp.$$" "$bin" 2>/dev/null; then
    rm -f "$marker"
    return 0
  fi
  rm -f "$bin.tmp.$$"
  printf '%s' "$fp" >"$marker" 2>/dev/null
  return 1
}

if stale; then
  command -v go >/dev/null 2>&1 || exit 0
  mkdir -p "$data" 2>/dev/null || exit 0
  if command -v flock >/dev/null 2>&1; then
    exec 9>"$bin.lock" || exit 0
    flock 9 || exit 0
    if stale; then
      build || exit 0
    fi
    exec 9>&-
  else
    build || exit 0
  fi
fi

printf '%s' "$stdin" | "$bin" "$mode"
exit 0
