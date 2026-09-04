#!/usr/bin/env bash
#
# Run one eval case by hand, against THIS checkout's canon.
#
# `claude plugin eval` is not a documented command yet, so a case is
# exercised in two separate `claude -p` invocations: one that plays the
# subject and one that grades it. They must stay separate — a session that
# has read graders/criteria.md writes a report that satisfies it.
#
#   ./evals/run-manual.sh go-conventions canon-new-binary
#   ./evals/run-manual.sh github-conventions gh-pin-check -o /tmp/evalrun
#
# A case that needs more than Read/Grep/Glob lists the extra tools, one per
# line, in <case>/allowed-tools.
set -euo pipefail

ROOT=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd -P)

MODEL=""
OUT_DIR=""
PLUGIN=""
CASE=""
SUBJECT_ONLY=0

usage() {
	cat <<-USAGE
		usage: run-manual.sh <plugin> <case> [-m MODEL] [-o OUT_DIR] [-s]

		  -m MODEL    model for both passes (default: the CLI's default)
		  -o OUT_DIR  where to write the report and grade (default: a temp dir)
		  -s          subject pass only; skip grading

		plugins: $(cd "$ROOT/plugins" && printf '%s ' */)
	USAGE
}

while [ $# -gt 0 ]; do
	case "$1" in
	-m) MODEL=${2:?-m needs a model}; shift 2 ;;
	-o) OUT_DIR=${2:?-o needs a directory}; shift 2 ;;
	-s) SUBJECT_ONLY=1; shift ;;
	-h | --help) usage; exit 0 ;;
	-*) echo "unknown flag: $1" >&2; usage >&2; exit 2 ;;
	*)
		if [ -z "$PLUGIN" ]; then PLUGIN=$1; elif [ -z "$CASE" ]; then CASE=$1; else echo "unexpected argument: $1" >&2; exit 2; fi
		shift ;;
	esac
done

[ -n "$PLUGIN" ] && [ -n "$CASE" ] || { usage >&2; exit 2; }

PLUGIN_ROOT="$ROOT/plugins/$PLUGIN"
CASE_DIR="$PLUGIN_ROOT/evals/$CASE"
PROMPT="$CASE_DIR/prompt.md"
CRITERIA="$CASE_DIR/graders/criteria.md"
[ -d "$PLUGIN_ROOT" ] || { echo "no such plugin: $PLUGIN" >&2; exit 2; }
[ -f "$PROMPT" ] || { echo "no such case: $CASE (missing $PROMPT)" >&2; exit 2; }

OUT_DIR=${OUT_DIR:-$(mktemp -d -t "eval-$PLUGIN-$CASE-XXXXXX")}
mkdir -p "$OUT_DIR"
REPORT="$OUT_DIR/report.md"
GRADE="$OUT_DIR/grade.md"

declare -a MODEL_ARGS=()
[ -n "$MODEL" ] && MODEL_ARGS=(--model "$MODEL")

declare -a ALLOWED_TOOLS=(Read Grep Glob)
if [ -f "$CASE_DIR/allowed-tools" ]; then
	while IFS= read -r tool; do
		[ -n "$tool" ] && ALLOWED_TOOLS+=("$tool")
	done < "$CASE_DIR/allowed-tools"
	# tools/run.sh dies when CLAUDE_PLUGIN_ROOT is unset; the checkout under
	# test is the canon this run exists to exercise.
	export CLAUDE_PLUGIN_ROOT="$PLUGIN_ROOT"
fi

# Prompts go in on stdin, never as the positional argument: --allowedTools and
# --add-dir are variadic, so a trailing prompt is parsed as more tool names.
subject_prompt() {
	cat <<-SUBJECT
		You are the SUBJECT of a regression test for the $PLUGIN plugin.
		Someone else grades your output.

		CANON SOURCE: do NOT use any installed or loaded $PLUGIN plugin and do
		NOT invoke its skills by name. The installed copy is a different release
		than the canon under test. Read the plugin under test from:

		  $PLUGIN_ROOT

		Start with the SKILL.md of the skill the task names (or the plugin's
		canon skill when it names none), then read the reference files its own
		routing table directs you to. Route honestly from what the task
		actually touches.

		HARD RULE: do not read any graders/ directory, any other eval case, or
		anything under docs/. Reading the grading criteria or the design intent
		would invalidate this run.

		Where the case names a command you cannot run, say what it would have
		checked and record it as a coverage gap rather than inventing a result.

		Your task follows. Follow it exactly. Your entire final answer is what it
		asks for — nothing added, no meta-commentary about being tested.

		$(cat "$PROMPT")
	SUBJECT
}

# The canon redirect is the whole point. A fresh session loads the INSTALLED
# plugin, which is whatever release is on disk — not this branch.
echo "==> subject pass: $PLUGIN/$CASE" >&2
subject_prompt | claude -p "${MODEL_ARGS[@]}" \
	--add-dir "$PLUGIN_ROOT" --allowedTools "${ALLOWED_TOOLS[@]}" | tee "$REPORT"

if [ "$SUBJECT_ONLY" = "1" ]; then
	echo "==> report: $REPORT" >&2
	exit 0
fi

[ -f "$CRITERIA" ] || { echo "no criteria at $CRITERIA; report left at $REPORT" >&2; exit 2; }

grader_prompt() {
	cat <<-GRADER
		Grade a report against an eval case's criteria. Be strict: the
		criteria are numbered conditions, and "close enough" is a fail. Answer
		each numbered condition PASS or FAIL with the evidence you relied on,
		check the Fail list too, then give one overall verdict line.

		=== CRITERIA ===
		$(cat "$CRITERIA")

		=== REPORT UNDER TEST ===
		$(cat "$REPORT")
	GRADER
}

# Separate invocation on purpose: the grader may see the criteria, the subject
# may not, and neither may be the session that wrote the canon.
echo "==> grading pass: $PLUGIN/$CASE" >&2
grader_prompt | claude -p "${MODEL_ARGS[@]}" --allowedTools Read | tee "$GRADE"

echo "==> report: $REPORT" >&2
echo "==> grade:  $GRADE" >&2
