#!/usr/bin/env bash
# End-to-end tests for the go-conventions hook launcher, hooks/goconv-hook.sh.
#
#   bash tests/goconv-hook-test.sh
#
# The Go module under hooks/goconv-hook/ has its own `testing` suite; this one
# drives the LAUNCHER, which is where the hook's fail-open behaviour, the
# pre-build short circuits, and the stdin plumbing live -- none of which the Go
# tests can see.
#
# The suite pins PATH to a shim holding only `go` and `gofmt`, so a machine
# with gofumpt installed still produces the message these assertions expect.
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
launcher="$root/plugins/go-conventions/hooks/goconv-hook.sh"
[ -f "$launcher" ] || {
  printf 'missing launcher: %s\n' "$launcher" >&2
  exit 1
}

tmp=$(mktemp -d "${TMPDIR:-/tmp}/goconv-hook-test.XXXXXX")
trap 'rm -rf "$tmp"' EXIT

shim="$tmp/bin"
mkdir -p "$shim"
for prog in go gofmt; do
  p=$(command -v "$prog") || {
    printf '%s is required and was not found on PATH\n' "$prog" >&2
    exit 1
  }
  ln -s "$p" "$shim/$prog"
done

export PATH="$shim:/usr/bin:/bin"
export CGO_ENABLED=0
export CLAUDE_PLUGIN_ROOT="$root/plugins/go-conventions"
# Nothing may land in the developer's real cache, whichever branch of the
# launcher's cache selection runs.
export XDG_CACHE_HOME="$tmp/xdg"
CLAUDE_PLUGIN_DATA=$(mktemp -d "$tmp/go-conventions-XXXXXX")
export CLAUDE_PLUGIN_DATA
# Set, and deliberately not a module: a bug that reaches for the fallback when
# stdin already carried a cwd shows up as silence rather than a wrong answer.
export CLAUDE_PROJECT_DIR="$tmp/not-a-module"
mkdir -p "$CLAUDE_PROJECT_DIR"

passed=0
failed=0

ok() {
  printf 'PASS %s\n' "$1"
  passed=$((passed + 1))
}

bad() {
  printf 'FAIL %s: %s\n' "$1" "$2"
  failed=$((failed + 1))
}

# Runs the launcher the way Claude Code does: the hook JSON on stdin, the mode
# as the only argument. Sets `out` and `status`.
run_hook() {
  status=0
  out=$(printf '%s' "$2" | bash "$launcher" "$1") || status=$?
}

expect_silent() {
  local name=$1
  if [ "$status" -ne 0 ]; then
    bad "$name" "exit status $status, want 0"
    return
  fi
  if [ -n "$out" ]; then
    bad "$name" "stdout is $out, want empty"
    return
  fi
  ok "$name"
}

badly_formatted="$tmp/work/bad.go"

test_format_rewrites_a_go_file() {
  local name=format_rewrites_a_go_file
  mkdir -p "$(dirname "$badly_formatted")"
  printf 'package work\nfunc  F( )  {}\n' >"$badly_formatted"

  run_hook format "{\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"$badly_formatted\"},\"cwd\":\"$tmp\"}"

  if [ "$status" -ne 0 ]; then
    bad "$name" "exit status $status, want 0"
    return
  fi
  case "$out" in
  *'"hookEventName":"PostToolUse"'*) ;;
  *)
    bad "$name" "stdout is $out, want a PostToolUse payload"
    return
    ;;
  esac
  case "$out" in
  *"gofmt rewrote $badly_formatted on disk; re-read it before the next Edit"*) ;;
  *)
    bad "$name" "stdout is $out, want the rewrite notice"
    return
    ;;
  esac
  if [ -n "$(gofmt -l "$badly_formatted")" ]; then
    bad "$name" "$badly_formatted is still not gofmt-clean"
    return
  fi
  ok "$name"
}

test_format_silent_when_already_clean() {
  local name=format_silent_when_already_clean
  run_hook format "{\"tool_name\":\"Edit\",\"tool_input\":{\"file_path\":\"$badly_formatted\"},\"cwd\":\"$tmp\"}"
  expect_silent "$name"
}

test_format_ignores_a_non_go_path() {
  local name=format_ignores_a_non_go_path
  local fresh
  fresh=$(mktemp -d "$tmp/go-conventions-fresh-XXXXXX")
  printf '#  heading\n' >"$tmp/work/notes.md"

  status=0
  out=$(printf '%s' "{\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"$tmp/work/notes.md\"},\"cwd\":\"$tmp\"}" |
    CLAUDE_PLUGIN_DATA="$fresh" bash "$launcher" format) || status=$?

  if [ "$status" -ne 0 ] || [ -n "$out" ]; then
    bad "$name" "exit status $status, stdout $out; want 0 and empty"
    return
  fi
  # The point of the pre-filter: a path the hook can do nothing with must not
  # cost a build.
  if [ -n "$(find "$fresh" -type f -print -quit)" ]; then
    bad "$name" "the launcher built into $fresh for a non-go path"
    return
  fi
  ok "$name"
}

test_session_start_reports_the_module() {
  local name=session_start_reports_the_module module="$tmp/module"
  mkdir -p "$module"
  printf 'module example.com/x\n\ngo 1.27\n' >"$module/go.mod"

  run_hook session-start "{\"cwd\":\"$module\"}"

  if [ "$status" -ne 0 ]; then
    bad "$name" "exit status $status, want 0"
    return
  fi
  case "$out" in
  *'"hookEventName":"SessionStart"'*'Go module example.com/x (go 1.27): load go-conventions:go-conventions before writing Go'*)
    ok "$name"
    ;;
  *) bad "$name" "stdout is $out, want the module notice" ;;
  esac
}

test_session_start_silent_without_go_mod() {
  local name=session_start_silent_without_go_mod plain="$tmp/plain"
  mkdir -p "$plain"
  run_hook session-start "{\"cwd\":\"$plain\"}"
  expect_silent "$name"
}

test_session_start_silent_on_a_malformed_module_line() {
  local name=session_start_silent_on_a_malformed_module_line module="$tmp/spacey"
  mkdir -p "$module"
  printf 'module example.com/x y\n\ngo 1.27\n' >"$module/go.mod"
  run_hook session-start "{\"cwd\":\"$module\"}"
  expect_silent "$name"
}

# A cwd the pre-filter cannot read literally must fall through to the binary,
# which unescapes the JSON properly -- declining on a value the grep got wrong
# would be a silent false negative.
test_session_start_reads_an_escaped_cwd() {
  local name=session_start_reads_an_escaped_cwd module="$tmp/esc\"quote" escaped
  mkdir -p "$module"
  printf 'module example.com/escaped\n\ngo 1.27\n' >"$module/go.mod"
  escaped=${module//\"/\\\"}

  run_hook session-start "{\"cwd\":\"$escaped\"}"

  if [ "$status" -ne 0 ]; then
    bad "$name" "exit status $status, want 0"
    return
  fi
  case "$out" in
  *'Go module example.com/escaped (go 1.27)'*) ok "$name" ;;
  *) bad "$name" "stdout is $out, want the module notice" ;;
  esac
}

# Exit 2 from a PostToolUse hook is Claude Code's BLOCKING status: a binary
# that panics must not take the session's Write and Edit calls down with it.
test_a_failing_binary_does_not_block_the_tool_call() {
  local name=a_failing_binary_does_not_block_the_tool_call stub
  stub=$(mktemp -d "$tmp/go-conventions-stub-XXXXXX")
  mkdir -p "$stub/hooks"
  printf '#!/bin/sh\nprintf "boom\\n" >&2\nexit 2\n' >"$stub/hooks/goconv-hook"
  chmod +x "$stub/hooks/goconv-hook"

  status=0
  out=$(printf '%s' "{\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"$tmp/work/bad.go\"},\"cwd\":\"$tmp\"}" |
    CLAUDE_PLUGIN_DATA="$stub" bash "$launcher" format 2>/dev/null) || status=$?

  if [ "$status" -ne 0 ]; then
    bad "$name" "exit status $status, want 0 -- 2 blocks the tool call"
    return
  fi
  ok "$name"
}

test_kill_switch_silences_the_hook() {
  local name=kill_switch_silences_the_hook
  printf 'package work\nfunc  G( )  {}\n' >"$tmp/work/switch.go"

  status=0
  out=$(printf '%s' "{\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"$tmp/work/switch.go\"},\"cwd\":\"$tmp\"}" |
    GOCONV_HOOK=off bash "$launcher" format) || status=$?

  if [ -z "$(gofmt -l "$tmp/work/switch.go")" ]; then
    bad "$name" "the hook formatted the file with GOCONV_HOOK=off"
    return
  fi
  expect_silent "$name"
}

test_format_rewrites_a_go_file
test_format_silent_when_already_clean
test_format_ignores_a_non_go_path
test_session_start_reports_the_module
test_session_start_silent_without_go_mod
test_session_start_silent_on_a_malformed_module_line
test_session_start_reads_an_escaped_cwd
test_a_failing_binary_does_not_block_the_tool_call
test_kill_switch_silences_the_hook

printf '\n%d passed, %d failed\n' "$passed" "$failed"
[ "$failed" -eq 0 ]
