# AGENTS.md

Notes for agents and humans changing this repository. This repo is a Claude
Code plugin marketplace; the shipped product is two plugins,
`plugins/go-conventions` and `plugins/github-conventions`, whose "code" is
mostly instruction prose that steers a model. A prose defect here is a real
defect.

The contribution mechanics — Conventional Commits, never bumping a plugin
version by hand, and the checks CI runs — live in `README.md` under
"Development". This file does not restate them.

## Where the rules live

Every rule has exactly one normative statement.

- `plugins/github-conventions/skills/github-conventions/SKILL.md` owns the
  precedence ladder both plugins share and its own routing table; each file
  under its `references/` owns the rules of one area (new repos, workflows,
  security, commits, pull requests).
- `plugins/go-conventions/skills/go-conventions/SKILL.md` owns the one rung
  the Go canon adds to that ladder and its routing table; each file under
  its `references/` owns one area. Where a Go rule touches repository
  hygiene, the Go reference points at the github-conventions reference and
  states nothing of its own.
- A file under a skill's `templates/` is the enforced form of a rule. Its
  header comment names the reference that owns the rule; the template
  carries no rule of its own.
- `agents/*.md` carry no rules. Each names the skill file it runs under and
  nothing else; prose there that explains a rule is a finding.

Rationale stays with the rule it explains, wherever that rule lives.

## Sanctioned audience renderings

Some places re-explain a rule for a reader who has not installed the plugin
or has not loaded the skill. Each is a deliberate exception to the one-home
rule, bounded by this table: a rendering carries only what its row says.

| rendering | may carry |
| --- | --- |
| `README.md`, intro and skills tables | each skill's name, one line on what it does, and when it triggers; the install commands; the scope of each plugin |
| `README.md`, workflow diagrams | the phases, modes, gates, and fan-out of each procedural skill, drawn from its `SKILL.md` |
| `README.md`, interaction map | which skill hands off to which, and the go → github dependency |
| `.claude-plugin/marketplace.json`, every `description` | the plugin names and the areas each covers |
| `plugins/*/.claude-plugin/plugin.json`, `description` | the areas covered and the skill names |
| `plugins/*/skills/*/SKILL.md`, frontmatter `description` | triggering conditions only — never the procedure |
| `plugins/*/agents/*.md`, frontmatter `description` | the agent's job in a line and how it is dispatched |
| the SessionStart hook's `additionalContext` | the module path, the go directive, and the instruction to load the canon skill |
| a `templates/*.go` file's doc comments | what the rendered file does and the contract it keeps, for a reader in the target repository who has no plugin installed |
| `templates/CLAUDE-pointer.md` | the install line and repo facts the plugin cannot know (module path, binary name, env prefix); no rule |
| `templates/README.md`, Development section | the commit convention, the pin/actionlint reminder, and where the linter pin lives, stated as instructions to a contributor |
| `plugins/*/evals/README.md`, Guards column | one line per case naming the rule it guards and the failure it catches |
| `plugins/*/evals/*/graders/criteria.md` | the pass and fail conditions of the rule under test, restated so a grader can read them without the canon |

A rendering left behind when a rule it carries moves is DRIFTED, and it must
be rewritten in the same change that moves the rule. A rendering that reaches
past its row is RESTATED. Extend the table in the same change that extends a
rendering.

## Templates and this repository's own files

This repository lives by both canons it ships: its workflows, lint config,
Makefile, license, and commitlint config are renderings of the templates.
`.github/scripts/dogfood-sync.sh`, run by `validate.yml`, renders each
template with this repo's values and diffs it against the live copy. Edit
the template, then re-render; never edit the live copy alone.

Live workflows carry commit-SHA pins with version comments while templates
carry major tags, so `dogfood-sync.sh` normalizes the pins away before it
diffs. Why templates stay unpinned, and that a copying skill pins right
after, is
`plugins/github-conventions/skills/github-conventions/references/workflows.md`,
"Pinning".

## Skill workflows are documented in the README

Every skill that executes a **procedure** has a mermaid flowchart in
`README.md` under "Skill workflows", kept in sync with its `SKILL.md`.

A skill is a procedure if its `SKILL.md` has any of: numbered phases or
steps, a modes table, ordered gates, or agent fan-out. The test is
mechanical so the question is checkable rather than argued each time.
`go-converge`, `go-review`, `go-new-project`, `github-converge`, and
`github-new-repo` are procedures.

**Changing a skill's phases, modes, gates, or fan-out is not complete until
its diagram matches.** A stale diagram is the failure that actually happens:
the rule catches a missing diagram, this clause catches a stale one.

Adding a skill adds its diagram in the same PR. Removing one removes it.

### Reference skills are exempt

A skill with no entry point, no ordered steps, and no output contract — one
other skills consult rather than run — gets no workflow diagram. `go-conventions`
and `github-conventions` (the canon skills) are the exempt case: references
are read by trigger and skipped otherwise, and a report lands in the caller's
output contract, not theirs.

An exemption claimed for a skill that in fact has phases or modes is a review
blocker.

### What an exempt skill owes instead

It appears in the skill-interaction map, the README's one marketplace-shape
diagram. Adding, removing, splitting or renaming a skill, or changing a
cross-skill handoff, updates that map in the same PR.

### Diagram conventions

Match the existing diagrams rather than inventing a dialect:

- `flowchart TD`.
- Solid arrows are the skill's own pipeline; dashed (`-.->`) are invocations
  of another skill or of shared machinery, labelled with what is invoked.
- `[["..."]]` is a step that fans out as parallel agents; `["..."]` is an
  ordinary step; `{"..."}` is a decision. A decision naming the user is a
  human gate, and every human gate belongs on the diagram — the approval
  boundary is the thing a reader most needs to find.
- `<br/>` for line breaks. Avoid quotes inside labels; mermaid handles
  escaping inconsistently and the existing diagrams do without them.
- Draw what the `SKILL.md` says, not what it ought to say. A diagram that
  flatters the design is worse than none, because it will be believed.

### Verifying a diagram

GitHub renders these, so a syntax error ships as a broken page. `validate.yml`
renders the README with mermaid-cli; run the same check before pushing:

```sh
npx --yes @mermaid-js/mermaid-cli@11 -i README.md -o /tmp/readme-check.md
```

## Plugin tools

This is the normative statement of the launcher contract; a skill's own
`## Scripts` section lists only which tools that skill uses, each with its
invocation line and the fail-loud sentence.

Every tool is a Go binary under `${CLAUDE_PLUGIN_ROOT}/tools/cmd/`, invoked
through the launcher, which builds it on first use:

```sh
bash "${CLAUDE_PLUGIN_ROOT}/tools/run.sh" <tool> [args...]
```

The launcher fails loud — a non-zero exit is a real failure, never an empty
result. A hook launcher (`hooks/*.sh`) is the one exception and fails open:
a broken build must never brick the tool call or session it wraps.

Binaries live in the per-plugin data directory, never in
`CLAUDE_PLUGIN_ROOT`. `CLAUDE_PLUGIN_DATA` selects the cache when its
basename is this plugin's name (the model runs tools from its own shell, so
the variable may hold another plugin's directory); otherwise
`${XDG_CACHE_HOME:-$HOME/.cache}/conventions-claude/<plugin>` does. The
tools launcher stamps a checksum of the sources it built from and rebuilds
when it no longer matches; a hook launcher, which must stay small enough to
read at a glance, rebuilds on a source mtime newer than the binary.

Tool modules follow the Go canon they ship (cobra, log/slog, Ginkgo, the
house lint config found by walking up to the repo root). Hook binaries are
the stated carve-out: stdlib only, so the first build works offline.

A tool's package doc must carry a `Spec:` line naming the reference that
defines its output and a `Callers:` line naming the skills that invoke it,
so changing either obliges the other: adding a caller in a skill is half the
change.

## Design records

`docs/specs/*.md` record how a decision was reached, on the date it was
reached. They are non-normative: where one disagrees with the live rules,
the live rules win and the record is history rather than a finding. They may
state what was decided and what was rejected. They may not restate what a
decision produced — a restatement there drifts, and nothing is allowed to
catch it.
