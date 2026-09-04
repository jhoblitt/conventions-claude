#!/usr/bin/env bash
# This repository ships templates and also lives by them. Render each template
# with this repo's values and diff it against the live copy, so the two
# renderings of one file cannot drift: edit the template, then re-render.
#
# Live workflows carry SHA pins with version comments (pinact); templates
# carry major tags on purpose, so pins are normalized away before the diff.
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
      -e "s|{{MODULES}}|$MODULES|g" "$1"
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
check "$GH/check-breaking-footer.sh" .github/scripts/check-breaking-footer.sh
check "$GH/LICENSE" LICENSE
check "$GO/.golangci.yml" .golangci.yml
check "$GO/Makefile" Makefile

[ "$fail" -eq 0 ] && echo "templates and live copies agree"
exit "$fail"
