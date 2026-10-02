// Package closekw finds GitHub closing keywords that sit directly before an
// issue or pull request reference, in a PR's title and description, in the
// messages of its commits, and in the merge and squash messages GitHub
// renders from them.
package closekw

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// pattern matches a closing keyword, an optional colon, whitespace, then a
// cross-repository reference (owner/repo#N, or owner/repo/pull/N or
// owner/repo/issues/N bare or as a github.com URL over http or https, with or
// without www, optionally in an <autolink>) or a bare same-repository #N. The
// URL forms are included because whether they close is unverified, and an
// unverified close is reported rather than assumed safe.
//
// Owner and repository follow GitHub's name grammar, widened where a miss
// would cost a wrong close and a false positive only a confirmation: an owner
// is up to 39 letters, digits, hyphens, or underscores (Enterprise Managed
// User handles carry one), starting with a letter or digit; a repository is
// up to 100 letters, digits, '.', '_', or '-'.
var pattern = regexp.MustCompile(
	`(?i)\b(?:close[sd]?|fix(?:e[sd])?|resolve[sd]?)\b:?\s+` +
		`(?:<?(?:https?://(?:www\.)?github\.com/)?[a-z0-9][\w-]{0,38}/[\w.-]{1,100}(?:#|/pull/|/issues/)[0-9]+|#[0-9]+)\b`)

// Match is one closing keyword next to a reference.
type Match struct {
	Source string // "title", "description", "commit <sha12>", "merge message", or "squash message"
	Line   int    // 1-based, within Source
	Text   string // the keyword and reference, whitespace collapsed
}

func (m Match) String() string { return fmt.Sprintf("%s:%d: %s", m.Source, m.Line, m.Text) }

// Scan returns every match in text, attributed to source. A match may span a
// line break between keyword and reference; it is reported on the keyword's
// line.
func Scan(source, text string) []Match {
	var out []Match

	for _, loc := range pattern.FindAllStringIndex(text, -1) {
		out = append(out, Match{
			Source: source,
			Line:   strings.Count(text[:loc[0]], "\n") + 1,
			Text:   strings.Join(strings.Fields(text[loc[0]:loc[1]]), " "),
		})
	}

	return out
}

const (
	descriptionSource = "description"
	titleSource       = "title"
)

func commitSource(sha string) string { return "commit " + sha[:min(12, len(sha))] }

// scanReader scans everything r holds as the description. A read cut short by
// ctx is an error, never a clean scan of what arrived.
func scanReader(ctx context.Context, r io.Reader) ([]Match, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read description: %w", err)
	}

	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("read description: %w", err)
	}

	return Scan(descriptionSource, string(b)), nil
}

// scanCommits scans the message of every commit in revRange (anything git
// rev-list accepts, such as origin/main..HEAD) in the repository at dir,
// oldest first.
func scanCommits(ctx context.Context, dir, revRange string) ([]Match, error) {
	revs, err := run(ctx, dir, "git", "rev-list", "--reverse", revRange, "--")
	if err != nil {
		return nil, err
	}

	var out []Match

	sc := bufio.NewScanner(bytes.NewReader(revs))
	for sc.Scan() {
		sha := strings.TrimSpace(sc.Text())
		if sha == "" {
			continue
		}

		msg, err := run(ctx, dir, "git", "show", "-s", "--format=%B", sha)
		if err != nil {
			return nil, err
		}

		out = append(out, Scan(commitSource(sha), string(msg))...)
	}

	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read rev-list: %w", err)
	}

	return out, nil
}

type prView struct {
	Title               *string `json:"title"`
	Body                *string `json:"body"`
	URL                 *string `json:"url"`
	HeadRefOid          *string `json:"headRefOid"`
	HeadRefName         *string `json:"headRefName"`
	HeadRepositoryOwner *struct {
		Login string `json:"login"`
	} `json:"headRepositoryOwner"`
}

// commitsQuery pages through every commit of a pull request; gh api
// --paginate supplies $endCursor from pageInfo.
const commitsQuery = `query($owner:String!,$name:String!,$number:Int!,$endCursor:String){` +
	`repository(owner:$owner,name:$name){pullRequest(number:$number){headRefOid ` +
	`commits(first:100,after:$endCursor){totalCount pageInfo{hasNextPage endCursor} nodes{commit{oid message}}}}}}`

type commitsPage struct {
	Data struct {
		Repository *struct {
			PullRequest *struct {
				HeadRefOid string `json:"headRefOid"`
				Commits    struct {
					TotalCount int `json:"totalCount"`
					Nodes      []struct {
						Commit struct {
							OID     string `json:"oid"`
							Message string `json:"message"`
						} `json:"commit"`
					} `json:"nodes"`
				} `json:"commits"`
			} `json:"pullRequest"`
		} `json:"repository"`
	} `json:"data"`
}

type prCommit struct{ oid, message string }

// scanPR scans pull request pr's title, description, and the full message of
// every one of its commits as GitHub holds them, through gh, then the merge
// and squash messages GitHub would build from them, and returns the head
// commit the scan saw. repo, when set, is passed to gh pr view as --repo.
// The title is scanned because GitHub's default merge and squash messages
// carry it onto the default branch.
func scanPR(ctx context.Context, dir string, pr int, repo string) ([]Match, string, error) {
	args := []string{"pr", "view", strconv.Itoa(pr), "--json", "title,body,url,headRefOid,headRefName,headRepositoryOwner"}
	if repo != "" {
		args = append(args, "--repo", repo)
	}

	out, err := run(ctx, dir, "gh", args...)
	if err != nil {
		return nil, "", err
	}

	var view prView
	if err = json.Unmarshal(out, &view); err != nil {
		return nil, "", fmt.Errorf("parse gh pr view output: %w", err)
	}

	for _, f := range []struct {
		name    string
		missing bool
	}{
		{"title", view.Title == nil},
		{"body", view.Body == nil},
		{"url", view.URL == nil},
		{"headRefOid", view.HeadRefOid == nil || *view.HeadRefOid == ""},
		{"headRefName", view.HeadRefName == nil || *view.HeadRefName == ""},
		{"headRepositoryOwner", view.HeadRepositoryOwner == nil || view.HeadRepositoryOwner.Login == ""},
	} {
		if f.missing {
			return nil, "", fmt.Errorf("parse gh pr view output: no %s field", f.name)
		}
	}

	host, owner, name, err := prLocation(*view.URL, pr)
	if err != nil {
		return nil, "", err
	}

	commits, head, err := prCommits(ctx, dir, host, owner, name, pr)
	if err != nil {
		return nil, "", err
	}

	// The title and description, and the commits, are read in separate calls;
	// a push between them would scan one head and report another.
	if head != *view.HeadRefOid {
		return nil, "", fmt.Errorf("PR #%d head moved during the scan (%s, then %s); scan again",
			pr, *view.HeadRefOid, head)
	}

	matches := Scan(titleSource, *view.Title)
	matches = append(matches, Scan(descriptionSource, *view.Body)...)

	for _, c := range commits {
		matches = append(matches, Scan(commitSource(c.oid), c.message)...)
	}

	for _, m := range mergeMessages(pr, view.HeadRepositoryOwner.Login, *view.HeadRefName, *view.Title, *view.Body) {
		matches = appendNew(matches, scanJoined(mergeMessageSource, m))
	}

	matches = appendNew(matches, scanJoined(squashMessageSource, squashMessage(pr, *view.Title, commits)))

	return matches, *view.HeadRefOid, nil
}

const (
	mergeMessageSource  = "merge message"
	squashMessageSource = "squash message"
)

// part is one piece of a message GitHub renders: a field the PR carries (its
// title, a commit message, the head branch), or the fixed text GitHub puts
// between fields.
type part struct {
	text  string
	field bool
}

func lit(s string) part   { return part{text: s} }
func field(s string) part { return part{text: s, field: true} }

// mergeMessages renders the merge commit message GitHub builds for pr:
// "Merge pull request #N from OWNER/BRANCH", a blank line, then the title by
// default, or the description where the repository is set to use it.
func mergeMessages(pr int, owner, branch, title, body string) [][]part {
	head := []part{
		lit(fmt.Sprintf("Merge pull request #%d from ", pr)),
		field(owner), lit("/"), field(branch), lit("\n\n"),
	}

	return [][]part{
		append(append([]part{}, head...), field(title)),
		append(append([]part{}, head...), field(body)),
	}
}

// squashMessage renders GitHub's default squash commit message for pr: the
// title with " (#N)", a blank line, then each commit message as a "* " list
// item, the items separated by a blank line.
func squashMessage(pr int, title string, commits []prCommit) []part {
	parts := []part{field(title), lit(fmt.Sprintf(" (#%d)", pr))}

	for _, c := range commits {
		parts = append(parts, lit("\n\n* "), field(c.message))
	}

	return parts
}

// scanJoined scans the message parts join to and keeps only the matches no
// single field holds whole: a match inside one field is already reported
// from that field's own scan (or, for an owner or branch name, cannot occur,
// since neither holds whitespace).
func scanJoined(source string, parts []part) []Match {
	var (
		text  strings.Builder
		spans [][2]int
	)

	for _, p := range parts {
		if p.field {
			spans = append(spans, [2]int{text.Len(), text.Len() + len(p.text)})
		}

		text.WriteString(p.text)
	}

	joined := text.String()

	var out []Match

	for _, loc := range pattern.FindAllStringIndex(joined, -1) {
		inside := false

		for _, sp := range spans {
			if loc[0] >= sp[0] && loc[1] <= sp[1] {
				inside = true

				break
			}
		}

		if !inside {
			out = append(out, Match{
				Source: source,
				Line:   strings.Count(joined[:loc[0]], "\n") + 1,
				Text:   strings.Join(strings.Fields(joined[loc[0]:loc[1]]), " "),
			})
		}
	}

	return out
}

// appendNew appends each of add not already in ms, so the two merge message
// variants report a match they share once.
func appendNew(ms, add []Match) []Match {
	for _, a := range add {
		if !slices.Contains(ms, a) {
			ms = append(ms, a)
		}
	}

	return ms
}

// prLocation takes the host, owner, and repository from the URL gh pr view
// reported, so the commits are asked of the very PR gh resolved rather than
// of whatever host and remote gh api would pick on its own.
func prLocation(raw string, pr int) (host, owner, name string, err error) {
	bad := fmt.Errorf("parse gh pr view output: url %q is not https://HOST/OWNER/REPO/pull/%d", raw, pr)

	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.RawQuery != "" || u.Fragment != "" {
		return "", "", "", bad
	}

	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(parts) != 4 || parts[0] == "" || parts[1] == "" || parts[2] != "pull" || parts[3] != strconv.Itoa(pr) {
		return "", "", "", bad
	}

	return u.Host, parts[0], parts[1], nil
}

// prCommits pages through every commit of pull request pr with gh api graphql
// --paginate --slurp, and returns them with the PR head every page reported.
// It fails unless it collected exactly the totalCount GitHub reports, and
// unless every page names the same head and the same count: pages are
// separate requests, and a push between them changes both.
func prCommits(ctx context.Context, dir, host, owner, name string, pr int) ([]prCommit, string, error) {
	out, err := run(ctx, dir, "gh", "api", "graphql", "--paginate", "--slurp", "--hostname", host,
		"-f", "owner="+owner, "-f", "name="+name, "-F", "number="+strconv.Itoa(pr),
		"-f", "query="+commitsQuery)
	if err != nil {
		return nil, "", err
	}

	var pages []commitsPage
	if err = json.Unmarshal(out, &pages); err != nil {
		return nil, "", fmt.Errorf("parse gh api graphql output: %w", err)
	}

	if len(pages) == 0 {
		return nil, "", errors.New("parse gh api graphql output: no pages")
	}

	var (
		commits []prCommit
		head    string
		total   = -1
	)

	for _, p := range pages {
		if p.Data.Repository == nil || p.Data.Repository.PullRequest == nil {
			return nil, "", fmt.Errorf("parse gh api graphql output: no pull request %s/%s#%d", owner, name, pr)
		}

		pull := p.Data.Repository.PullRequest
		if pull.HeadRefOid == "" {
			return nil, "", errors.New("parse gh api graphql output: no headRefOid")
		}

		if head != "" && pull.HeadRefOid != head {
			return nil, "", fmt.Errorf("PR #%d head moved during the scan (%s, then %s); scan again",
				pr, head, pull.HeadRefOid)
		}

		head = pull.HeadRefOid

		c := pull.Commits
		if total != -1 && c.TotalCount != total {
			return nil, "", fmt.Errorf("PR #%d commit count changed during the scan (%d, then %d); scan again",
				pr, total, c.TotalCount)
		}

		total = c.TotalCount

		for _, n := range c.Nodes {
			if n.Commit.OID == "" {
				return nil, "", errors.New("parse gh api graphql output: a commit has no oid")
			}

			commits = append(commits, prCommit{oid: n.Commit.OID, message: n.Commit.Message})
		}
	}

	if len(commits) != total {
		return nil, "", fmt.Errorf("PR #%d: collected %d of %d commits; scan incomplete", pr, len(commits), total)
	}

	return commits, head, nil
}

func run(ctx context.Context, dir, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // G204: argv, never a shell string
	cmd.Dir = dir

	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}

	return out, nil
}
