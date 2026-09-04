#!/usr/bin/env bash
# Every references/<file>, templates/<file>, and ${CLAUDE_PLUGIN_ROOT}/<path>
# a skill, reference, or agent names must exist. A pointer to a file that is
# not there is a rule with no home.
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

fail=0

for skill in plugins/*/skills/*/; do
  for f in "$skill"SKILL.md "$skill"references/*.md; do
    [ -f "$f" ] || continue
    while IFS= read -r ref; do
      [ -e "$skill$ref" ] && continue
      # A pointer into another plugin's skill is written the same way.
      found=0
      for other in plugins/*/skills/*/"$ref"; do [ -e "$other" ] && found=1 && break; done
      [ "$found" -eq 1 ] && continue
      echo "$f: no such file $ref"
      fail=1
    done < <(grep -oE '(references|templates)/[A-Za-z0-9._-]+' "$f" | sort -u)
  done
done

for f in plugins/*/agents/*.md plugins/*/skills/*/SKILL.md plugins/*/skills/*/references/*.md; do
  [ -f "$f" ] || continue
  plugin=${f#plugins/}
  plugin=plugins/${plugin%%/*}
  while IFS= read -r ref; do
    path=${ref#\$\{CLAUDE_PLUGIN_ROOT\}/}
    [ -e "$plugin/$path" ] && continue
    echo "$f: no such file $plugin/$path"
    fail=1
  done < <(grep -oE '\$\{CLAUDE_PLUGIN_ROOT\}/[A-Za-z0-9._/-]+' "$f" | sort -u)
done

[ "$fail" -eq 0 ] && echo "every pointer resolves"
exit "$fail"
