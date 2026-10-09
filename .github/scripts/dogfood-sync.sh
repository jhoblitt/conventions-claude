#!/usr/bin/env bash
# This repository ships templates and also lives by them. Render each template
# with this repo's values and diff it against the live copy, so the two
# renderings of one file cannot drift: edit the template, then re-render.
#
# Live pins (a SHA plus its version comment, from pinact) are normalized to
# the major tag before the diff. On the template side only
# ossf/scorecard-action's full tag is normalized, and the check fails if that
# template stops carrying one; a full tag in any other template still shows
# up as drift. What tag a template carries is references/workflows.md,
# "Pinning".
# dependabot.yml is excluded: the live file appends gomod entries to the
# template's github-actions entry, which is a merge, not a rendering.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

GH=plugins/github-conventions/skills/github-conventions/templates
GO=plugins/go-conventions/skills/go-conventions/templates

OWNER=jhoblitt
REPO=conventions-claude
MODULE=github.com/jhoblitt/conventions-claude
CODEQL_LANGUAGES=go
CODEQL_BUILD_MODE=autobuild
MODULES="plugins/github-conventions/tools plugins/go-conventions/tools plugins/go-conventions/hooks/goconv-hook"

render() {
  sed -e "s|{{OWNER}}|$OWNER|g" \
      -e "s|{{REPO}}|$REPO|g" \
      -e "s|{{MODULE}}|$MODULE|g" \
      -e "s|{{CODEQL_LANGUAGES}}|$CODEQL_LANGUAGES|g" \
      -e "s|{{CODEQL_BUILD_MODE}}|$CODEQL_BUILD_MODE|g" \
      -e "s|{{MODULES}}|$MODULES|g" "$1" |
    sed -E 's/(uses: ossf\/scorecard-action)@(v[0-9]+)(\.[0-9]+)+$/\1@\2/'
}

unpin() {
  sed -E 's/@[0-9a-f]{40} # (v[0-9]+)(\.[0-9]+)*$/@\1/' "$1"
}

fail=0
check() {
  local template=$1 live=$2
  if [ ! -f "$live" ]; then
    echo "missing live copy: $live (rendered from $template)"
    fail=1
    return
  fi
  if ! diff -u --label "rendered $template" --label "$live" <(render "$template") <(unpin "$live"); then
    echo "drift: $live is not a rendering of $template"
    fail=1
  fi
}

for w in workflow-lint codeql dependency-review scorecard commitlint; do
  check "$GH/$w.yml" ".github/workflows/$w.yml"
done
check "$GH/.commitlintrc.yml" .commitlintrc.yml
check "$GH/breaking-footer/main.go" .github/tools/breaking-footer/main.go
check "$GH/LICENSE" LICENSE
check "$GO/.golangci.yml" .golangci.yml
check "$GO/Makefile" Makefile

if ! grep -Eq 'uses: ossf/scorecard-action@v[0-9]+\.[0-9]+\.[0-9]+$' "$GH/scorecard.yml"; then
  echo "$GH/scorecard.yml: ossf/scorecard-action needs a full vX.Y.Z tag (references/workflows.md, \"Pinning\")"
  fail=1
fi

[ "$fail" -eq 0 ] && echo "templates and live copies agree"
exit "$fail"
