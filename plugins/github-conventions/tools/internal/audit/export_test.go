package audit

import (
	"context"
	"log/slog"
)

// ParseSlug exposes the origin-URL parser to the spec in audit_test.
var ParseSlug = parseSlug

// APIFunc is the gh call the ruleset lookup makes; the spec substitutes its own.
type APIFunc = apiFunc

// RulesetDetail exposes the by-id expansion, so the spec can drive the fallback
// that keeps one unreadable ruleset from sinking the whole check.
func RulesetDetail(ctx context.Context, logger *slog.Logger, api APIFunc, list []Ruleset) []Ruleset {
	return ghRulesetLookup{logger: logger, api: api}.detail(ctx, "", "owner/repo", list)
}
