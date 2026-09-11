package audit

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"
)

// Ruleset is the part of a GitHub repository ruleset this audit reads.
type Ruleset struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Target      string     `json:"target"`
	Enforcement string     `json:"enforcement"`
	Conditions  Conditions `json:"conditions"`
	Rules       []Rule     `json:"rules"`
}

// Conditions narrows the refs a [Ruleset] applies to.
type Conditions struct {
	RefName RefName `json:"ref_name"`
}

// RefName is a ruleset's ref-name condition.
type RefName struct {
	Include []string `json:"include"`
	Exclude []string `json:"exclude"`
}

// Rule is one protection a [Ruleset] enforces.
type Rule struct {
	Type string `json:"type"`
}

func rulesetRow(ctx context.Context, opts Options) Row {
	const (
		canon = "an active branch ruleset on ~DEFAULT_BRANCH with deletion and non_fast_forward rules"
		fix   = "gh api --method POST repos/{owner}/{repo}/rulesets --input templates/ruleset.json"
	)

	if !opts.Remote {
		return skippedRow("ruleset", "default-branch", canon)
	}

	rulesets, err := opts.remoteLookup().Rulesets(ctx, opts.Dir)
	if err != nil {
		opts.Logger.WarnContext(ctx, "ruleset lookup failed", slog.Any("error", err))

		return gapRow("ruleset", "default-branch", "gh api failed: "+err.Error(), canon, fix)
	}

	summaries := make([]string, 0, len(rulesets))
	for i := range rulesets {
		if protectsDefaultBranch(&rulesets[i]) {
			return okRow("ruleset", "default-branch", summarize(&rulesets[i]), canon)
		}
		summaries = append(summaries, summarize(&rulesets[i]))
	}

	return gapRow("ruleset", "default-branch",
		listing("rulesets", summaries, "no rulesets on the repository"), canon, fix)
}

func protectsDefaultBranch(rs *Ruleset) bool {
	if rs.Enforcement != "active" || rs.Target != "branch" {
		return false
	}
	if !slices.Contains(rs.Conditions.RefName.Include, "~DEFAULT_BRANCH") {
		return false
	}
	types := ruleTypes(rs)

	return slices.Contains(types, "deletion") && slices.Contains(types, "non_fast_forward")
}

func summarize(rs *Ruleset) string {
	return fmt.Sprintf("%s: %s %s ruleset on %s with %s",
		rs.Name, rs.Enforcement, rs.Target,
		join(rs.Conditions.RefName.Include, "no refs"),
		join(ruleTypes(rs), "no rules"))
}

func ruleTypes(rs *Ruleset) []string {
	types := make([]string, 0, len(rs.Rules))
	for _, rule := range rs.Rules {
		types = append(types, rule.Type)
	}

	return types
}

func join(items []string, empty string) string {
	if len(items) == 0 {
		return empty
	}

	return strings.Join(items, ", ")
}

// Rulesets implements [RemoteLookup] against the GitHub API.
func (l ghRemoteLookup) Rulesets(ctx context.Context, dir string) ([]Ruleset, error) {
	slug, err := originSlug(ctx, dir)
	if err != nil {
		return nil, err
	}

	var list []Ruleset
	if err := l.api(ctx, dir, &list, "--paginate", "repos/"+slug+"/rulesets"); err != nil {
		return nil, err
	}

	return l.detail(ctx, dir, slug, list), nil
}

// detail refetches each ruleset by id, the documented source of conditions and
// rules; the list endpoint carries neither.
func (l ghRemoteLookup) detail(ctx context.Context, dir, slug string, list []Ruleset) []Ruleset {
	rulesets := make([]Ruleset, 0, len(list))
	for i := range list {
		var full Ruleset
		if err := l.api(ctx, dir, &full, fmt.Sprintf("repos/%s/rulesets/%d", slug, list[i].ID)); err != nil {
			// The list defaults to includes_parents=true, so it names org-level
			// rulesets this token may not be able to read on its own. Keeping
			// the list entry leaves the check short of conditions and rules, so
			// it can only report a gap, never a false pass.
			l.logger.WarnContext(ctx, "ruleset detail unavailable",
				slog.Int64("ruleset_id", list[i].ID), slog.Any("error", err))
			rulesets = append(rulesets, list[i])

			continue
		}
		rulesets = append(rulesets, full)
	}

	return rulesets
}
