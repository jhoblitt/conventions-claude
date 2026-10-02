package closekw_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jhoblitt/conventions-claude/plugins/github-conventions/tools/internal/closekw"
)

var _ = Describe("Scan", func() {
	texts := func(ms []closekw.Match) []string {
		out := make([]string, 0, len(ms))
		for _, m := range ms {
			out = append(out, m.Text)
		}

		return out
	}

	DescribeTable("reports a keyword directly before a reference",
		func(text, want string) {
			Expect(texts(closekw.Scan("description", text))).To(ConsistOf(want))
		},
		Entry("cross-repo, prose naming a fix", "the draft fix ceph/ceph#72250 landed", "fix ceph/ceph#72250"),
		Entry("bare same-repo", "Fixes #25", "Fixes #25"),
		Entry("trailing colon", "Resolves: #7", "Resolves: #7"),
		Entry("past tense", "closed o/r#3", "closed o/r#3"),
		Entry("pull URL", "fixes https://github.com/o/r/pull/9", "fixes https://github.com/o/r/pull/9"),
		Entry("issue URL", "close https://github.com/o/r/issues/9", "close https://github.com/o/r/issues/9"),
		Entry("any case", "FIXED #1", "FIXED #1"),
		Entry("line break between", "resolve\n#4", "resolve #4"),
		Entry("bare pull path", "fixes o/r/pull/9", "fixes o/r/pull/9"),
		Entry("http URL", "fixes http://github.com/o/r/pull/9", "fixes http://github.com/o/r/pull/9"),
		Entry("www URL", "fixes https://www.github.com/o/r/issues/9", "fixes https://www.github.com/o/r/issues/9"),
		Entry("autolinked URL", "fixes <https://github.com/o/r/pull/9>", "fixes <https://github.com/o/r/pull/9"),
		Entry("underscore owner, as an EMU handle", "fixes name_shortcode/r#1", "fixes name_shortcode/r#1"),
	)

	DescribeTable("ignores a reference no keyword precedes",
		func(text string) {
			Expect(closekw.Scan("description", text)).To(BeEmpty())
		},
		Entry("reworded", "the fix is ceph/ceph#72250"),
		Entry("parenthesised", "a fix (ceph/ceph#72250)"),
		Entry("keyword inside a word", "prefixes #3 and suffixed #4"),
		Entry("keyword without a reference", "fixes the parser"),
		Entry("other tracker URL", "fixes https://tracker.ceph.com/issues/80948"),
		Entry("owner starting with a hyphen", "fixes -ab/r#1"),
		Entry("owner with a dot", "fixes a.b/r#1"),
		Entry("owner past 39 characters", "fixes "+strings.Repeat("a", 40)+"/r#1"),
		Entry("repository past 100 characters", "fixes o/"+strings.Repeat("r", 101)+"#1"),
	)

	It("accepts names at GitHub's length limits", func() {
		text := "fixes " + strings.Repeat("a", 39) + "/" + strings.Repeat("r", 100) + "#1"
		Expect(closekw.Scan("description", text)).To(HaveLen(1))
	})

	It("reports each match with its source and line", func() {
		ms := closekw.Scan("commit abc", "subject\n\nbody\nFixes #2 and closes o/r#3\n")
		Expect(ms).To(HaveLen(2))
		Expect(ms[0].String()).To(Equal("commit abc:4: Fixes #2"))
		Expect(ms[1].String()).To(Equal("commit abc:4: closes o/r#3"))
	})
})

var _ = Describe("Run", func() {
	var stdout, stderr *bytes.Buffer

	BeforeEach(func() {
		stdout, stderr = &bytes.Buffer{}, &bytes.Buffer{}
	})

	run := func(ctx SpecContext, stdin string, args ...string) error {
		return closekw.Run(ctx, args, strings.NewReader(stdin), stdout, stderr)
	}

	It("refuses to run with nothing to scan", func(ctx SpecContext) {
		Expect(run(ctx, "")).To(MatchError(ContainSubstring("nothing to scan")))
	})

	It("scans a description read from stdin", func(ctx SpecContext) {
		Expect(run(ctx, "draft fix o/r#1\n", "--description", "-")).To(Succeed())
		Expect(stdout.String()).To(Equal("description:1: fix o/r#1\n1 matches\n"))
	})

	It("scans a description read from a file", func(ctx SpecContext) {
		path := filepath.Join(GinkgoT().TempDir(), "body.md")
		Expect(os.WriteFile(path, []byte("ok\nnothing here\n"), 0o600)).To(Succeed())

		Expect(run(ctx, "", "--description", path)).To(Succeed())
		Expect(stdout.String()).To(Equal("0 matches\n"))
	})

	It("fails loud on a description it cannot open", func(ctx SpecContext) {
		Expect(run(ctx, "", "--description", filepath.Join(GinkgoT().TempDir(), "missing"))).
			To(MatchError(ContainSubstring("open description")))
	})

	Context("over a git range", func() {
		var dir string

		gitIn := func(args ...string) string {
			cmd := exec.Command("git", args...)
			cmd.Dir = dir
			cmd.Env = append(os.Environ(),
				"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null",
				"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@example.com",
				"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@example.com")
			out, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred(), string(out))

			return strings.TrimSpace(string(out))
		}

		BeforeEach(func() {
			dir = GinkgoT().TempDir()
			gitIn("init", "-q", "-b", "main")
			gitIn("commit", "-q", "--allow-empty", "-m", "base: fixes #1")
			gitIn("commit", "-q", "--allow-empty", "-m", "feat: a\n\nFixes #25")
			gitIn("commit", "-q", "--allow-empty", "-m", "docs: b\n\nthe draft fix o/r#9")
		})

		It("scans each commit's message in the range, oldest first", func(ctx SpecContext) {
			first := gitIn("rev-parse", "--short=12", "HEAD~1")
			second := gitIn("rev-parse", "--short=12", "HEAD")

			Expect(run(ctx, "", "--dir", dir, "--range", "HEAD~2..HEAD")).To(Succeed())
			Expect(stdout.String()).To(Equal(
				"commit " + first + ":3: Fixes #25\n" +
					"commit " + second + ":3: fix o/r#9\n" +
					"2 matches\n"))
		})

		It("fails loud on a range git rejects", func(ctx SpecContext) {
			Expect(run(ctx, "", "--dir", dir, "--range", "nope..HEAD")).
				To(MatchError(ContainSubstring("git rev-list")))
		})
	})

	It("never reports success for a run cut short", func(ctx SpecContext) {
		cut, cancel := context.WithCancel(ctx)
		cancel()

		err := closekw.Run(cut, []string{"--description", "-"}, strings.NewReader("Fixes #1\n"), stdout, stderr)
		Expect(err).To(MatchError(context.Canceled))
		Expect(stdout.String()).To(BeEmpty())
	})

	DescribeTable("rejects a flag combination it cannot honor",
		func(ctx SpecContext, want string, args ...string) {
			Expect(run(ctx, "", args...)).To(MatchError(ContainSubstring(want)))
		},
		Entry("--pr with --description", "none of the others can be", "--pr", "1", "--description", "-"),
		Entry("--pr with --range", "none of the others can be", "--pr", "1", "--range", "a..b"),
		Entry("--repo without --pr", "--repo needs --pr", "--repo", "o/r", "--range", "a..b"),
		Entry("a non-positive --pr", "positive number", "--pr", "0"),
	)

	Context("over a pull request, through gh", func() {
		var argsFile, viewFile, pagesFile string

		// The fake logs each invocation's argv on one line. gh pr view answers
		// from viewFile and gh api from pagesFile; FAKE_GH_MODE "fail" makes
		// gh pr view fail, "bad" makes it print non-JSON, and "apifail" makes
		// gh api fail.
		BeforeEach(func() {
			bin := GinkgoT().TempDir()
			argsFile = filepath.Join(bin, "args")
			viewFile = filepath.Join(bin, "view.json")
			pagesFile = filepath.Join(bin, "pages.json")

			script := "#!/bin/sh\n" +
				"echo \"$*\" >> \"$FAKE_GH_ARGS\"\n" +
				"if [ \"$1\" = api ]; then\n" +
				"  [ \"$FAKE_GH_MODE\" = apifail ] && { echo 'HTTP 502: Bad Gateway' >&2; exit 1; }\n" +
				"  cat \"$FAKE_GH_PAGES\"; exit 0\n" +
				"fi\n" +
				"case \"$FAKE_GH_MODE\" in\n" +
				"fail) echo 'GraphQL: Could not resolve to a PullRequest' >&2; exit 1 ;;\n" +
				"bad) echo '{not json' ;;\n" +
				"*) cat \"$FAKE_GH_VIEW\" ;;\n" +
				"esac\n"
			Expect(os.WriteFile(filepath.Join(bin, "gh"), []byte(script), 0o700)).To(Succeed())

			GinkgoT().Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			GinkgoT().Setenv("FAKE_GH_ARGS", argsFile)
			GinkgoT().Setenv("FAKE_GH_VIEW", viewFile)
			GinkgoT().Setenv("FAKE_GH_PAGES", pagesFile)
			GinkgoT().Setenv("FAKE_GH_MODE", "ok")
		})

		writeJSON := func(path string, v any) {
			b, err := json.Marshal(v)
			Expect(err).NotTo(HaveOccurred())
			Expect(os.WriteFile(path, b, 0o600)).To(Succeed())
		}

		// writeView serves a gh pr view answer for a well-formed PR #26 on
		// github.com whose head is "h2", with fields replaced from set (a nil
		// value drops the field, a jsonNull value sends null).
		writeView := func(set map[string]any) {
			view := map[string]any{
				"title":               "a title",
				"body":                "",
				"url":                 "https://github.com/o/r/pull/26",
				"headRefOid":          "h2",
				"headRefName":         "topic",
				"headRepositoryOwner": map[string]any{"login": "fork"},
			}
			for k, v := range set {
				if v == nil {
					delete(view, k)
				} else {
					view[k] = v
				}
			}

			writeJSON(viewFile, view)
		}

		node := func(oid, message string) map[string]any {
			return map[string]any{"commit": map[string]any{"oid": oid, "message": message}}
		}

		// pageAt is one GraphQL page reporting head as the PR head; page is
		// one reporting the head writeView serves.
		pageAt := func(head string, total int, nodes ...map[string]any) map[string]any {
			return map[string]any{"data": map[string]any{"repository": map[string]any{"pullRequest": map[string]any{
				"headRefOid": head,
				"commits":    map[string]any{"totalCount": total, "nodes": nodes},
			}}}}
		}

		page := func(total int, nodes ...map[string]any) map[string]any { return pageAt("h2", total, nodes...) }

		writePages := func(pages ...any) { writeJSON(pagesFile, pages) }

		invocations := func() []string {
			b, err := os.ReadFile(argsFile)
			Expect(err).NotTo(HaveOccurred())

			return strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
		}

		graphqlCall := func(host, owner, name string) string {
			return "api graphql --paginate --slurp --hostname " + host +
				" -f owner=" + owner + " -f name=" + name + " -F number=26 -f query="
		}

		It("scans the title, the description, and every commit's full message", func(ctx SpecContext) {
			writeView(map[string]any{"title": "closes #7 for good", "body": "Motivation.\n\nthe draft fix o/r#9"})
			writePages(
				page(3, node("0123456789abcdef0123", "feat: a\n\nFixes #25")),
				page(3,
					node("1111111111111111aaaa", "docs: a subject long enough that the API headline would cut it before fix #3"),
					node("h2", "chore: c")),
			)

			Expect(run(ctx, "", "--pr", "26", "--repo", "o/r")).To(Succeed())
			Expect(stdout.String()).To(Equal(
				"title:1: closes #7\n" +
					"description:3: fix o/r#9\n" +
					"commit 0123456789ab:3: Fixes #25\n" +
					"commit 111111111111:1: fix #3\n" +
					"head h2\n" +
					"4 matches\n"))
		})

		It("reports a keyword and reference the merge message joins across the branch and the title", func(ctx SpecContext) {
			writeView(map[string]any{"headRefName": "x/fixes", "title": "#5 tidy"})
			writePages(page(1, node("h2", "chore: c")))

			Expect(run(ctx, "", "--pr", "26")).To(Succeed())
			Expect(stdout.String()).To(Equal("merge message:1: fixes #5\nhead h2\n1 matches\n"))
		})

		It("reports a keyword and reference the merge message joins across the branch and the description", func(ctx SpecContext) {
			writeView(map[string]any{"headRefName": "x/fixes", "body": "#5 is the follow-up"})
			writePages(page(1, node("h2", "chore: c")))

			Expect(run(ctx, "", "--pr", "26")).To(Succeed())
			Expect(stdout.String()).To(Equal("merge message:1: fixes #5\nhead h2\n1 matches\n"))
		})

		It("renders the squash message as GitHub does, with each commit a list item", func(ctx SpecContext) {
			// GitHub puts "* " before each squashed commit, so a commit ending
			// "fixes" and the next starting "#5" are joined with "* " between
			// them, which is no closing keyword.
			writeView(nil)
			writePages(page(2, node("h1", "feat: a\n\nthis fixes"), node("h2", "#5 next")))

			Expect(run(ctx, "", "--pr", "26")).To(Succeed())
			Expect(stdout.String()).To(Equal("head h2\n0 matches\n"))
		})

		It("does not report a match one field holds again from a rendered message", func(ctx SpecContext) {
			writeView(map[string]any{"title": "closes #7", "body": "resolves o/r#8"})
			writePages(page(1, node("h2", "feat: a\n\nFixes #9")))

			Expect(run(ctx, "", "--pr", "26")).To(Succeed())
			Expect(stdout.String()).To(Equal(
				"title:1: closes #7\n" +
					"description:1: resolves o/r#8\n" +
					"commit h2:3: Fixes #9\n" +
					"head h2\n" +
					"3 matches\n"))
		})

		It("asks for the commits of the PR gh resolved, on its host", func(ctx SpecContext) {
			writeView(map[string]any{"url": "https://ghe.example/team/proj/pull/26"})
			writePages(page(1, node("h2", "x")))

			Expect(run(ctx, "", "--pr", "26")).To(Succeed())

			calls := invocations()
			Expect(calls).To(HaveLen(2))
			Expect(calls[0]).To(Equal("pr view 26 --json title,body,url,headRefOid,headRefName,headRepositoryOwner"))
			Expect(calls[1]).To(HavePrefix(graphqlCall("ghe.example", "team", "proj")))

			query := strings.TrimPrefix(calls[1], graphqlCall("ghe.example", "team", "proj"))
			Expect(query).To(HavePrefix("query($owner:String!,$name:String!,$number:Int!,$endCursor:String){"))
			Expect(query).To(ContainSubstring("pullRequest(number:$number){headRefOid commits(first:100,after:$endCursor){"))
			Expect(query).To(ContainSubstring("totalCount pageInfo{hasNextPage endCursor} nodes{commit{oid message}}"))
		})

		It("passes --repo through to gh pr view only", func(ctx SpecContext) {
			writeView(nil)
			writePages(page(1, node("h2", "x")))

			Expect(run(ctx, "", "--pr", "26", "--repo", "o/r")).To(Succeed())

			calls := invocations()
			Expect(calls[0]).To(Equal("pr view 26 --json title,body,url,headRefOid,headRefName,headRepositoryOwner --repo o/r"))
			Expect(calls[1]).To(HavePrefix(graphqlCall("github.com", "o", "r")))
		})

		It("reports the head it scanned, so the merge can be bound to it", func(ctx SpecContext) {
			writeView(nil)
			writePages(page(2, node("h1", "a")), page(2, node("h2", "b")))

			Expect(run(ctx, "", "--pr", "26")).To(Succeed())
			Expect(stdout.String()).To(Equal("head h2\n0 matches\n"))
		})

		It("fails loud when the head moved between gh pr view and the commits", func(ctx SpecContext) {
			writeView(nil)
			writePages(pageAt("h3", 1, node("h3", "a")))

			Expect(run(ctx, "", "--pr", "26")).To(MatchError(ContainSubstring("head moved during the scan (h2, then h3)")))
			Expect(stdout.String()).To(BeEmpty())
		})

		It("fails loud when the head moved between pages", func(ctx SpecContext) {
			// The last page agrees with gh pr view, so only the per-page check
			// can catch the first one.
			writeView(nil)
			writePages(pageAt("h3", 2, node("h1", "a")), pageAt("h2", 2, node("h2", "b")))

			Expect(run(ctx, "", "--pr", "26")).To(MatchError(ContainSubstring("head moved during the scan (h3, then h2)")))
			Expect(stdout.String()).To(BeEmpty())
		})

		It("takes the head from GitHub, not from the order of the commits", func(ctx SpecContext) {
			writeView(nil)
			writePages(page(2, node("h2", "b")), page(2, node("h1", "fixes #4")))

			Expect(run(ctx, "", "--pr", "26")).To(Succeed())
			Expect(stdout.String()).To(Equal("commit h1:1: fixes #4\nhead h2\n1 matches\n"))
		})

		It("fails loud when the pages hold fewer commits than the PR has", func(ctx SpecContext) {
			nodes := make([]map[string]any, 100)
			for i := range nodes {
				nodes[i] = node(fmt.Sprintf("%040d", i), "c")
			}

			writeView(nil)
			writePages(page(143, nodes...))

			Expect(run(ctx, "", "--pr", "26")).
				To(MatchError("PR #26: collected 100 of 143 commits; scan incomplete"))
			Expect(stdout.String()).To(BeEmpty())
		})

		It("fails loud when the pages hold more commits than the PR has", func(ctx SpecContext) {
			writeView(nil)
			writePages(page(1, node("h1", "a"), node("h2", "b")))

			Expect(run(ctx, "", "--pr", "26")).
				To(MatchError("PR #26: collected 2 of 1 commits; scan incomplete"))
		})

		It("fails loud when the commit count changes between pages", func(ctx SpecContext) {
			writeView(nil)
			writePages(page(2, node("h1", "a")), page(3, node("h2", "b")))

			Expect(run(ctx, "", "--pr", "26")).To(MatchError(ContainSubstring("commit count changed during the scan")))
		})

		DescribeTable("fails loud on a PR url it cannot place",
			func(ctx SpecContext, url string) {
				writeView(map[string]any{"url": url})

				Expect(run(ctx, "", "--pr", "26")).
					To(MatchError(ContainSubstring("is not https://HOST/OWNER/REPO/pull/26")))
				Expect(stdout.String()).To(BeEmpty())
			},
			Entry("not https", "http://github.com/o/r/pull/26"),
			Entry("no host", "https:///o/r/pull/26"),
			Entry("an empty owner", "https://github.com//r/pull/26"),
			Entry("an issue, not a pull", "https://github.com/o/r/issues/26"),
			Entry("another PR's number", "https://github.com/o/r/pull/27"),
			Entry("a path too short", "https://github.com/o/pull/26"),
			Entry("a path too long", "https://github.com/o/r/pull/26/files"),
			Entry("a query", "https://github.com/o/r/pull/26?x=1"),
			Entry("a fragment", "https://github.com/o/r/pull/26#top"),
			Entry("not a URL", "::"),
		)

		It("fails loud with gh's error when gh pr view fails", func(ctx SpecContext) {
			GinkgoT().Setenv("FAKE_GH_MODE", "fail")

			err := run(ctx, "", "--pr", "26")
			Expect(err).To(MatchError(ContainSubstring("Could not resolve to a PullRequest")))
			Expect(stdout.String()).To(BeEmpty())
		})

		It("fails loud with gh's error when the commits cannot be read", func(ctx SpecContext) {
			writeView(nil)
			GinkgoT().Setenv("FAKE_GH_MODE", "apifail")

			Expect(run(ctx, "", "--pr", "26")).To(MatchError(ContainSubstring("HTTP 502: Bad Gateway")))
			Expect(stdout.String()).To(BeEmpty())
		})

		It("fails loud on gh pr view output that is not JSON", func(ctx SpecContext) {
			GinkgoT().Setenv("FAKE_GH_MODE", "bad")

			Expect(run(ctx, "", "--pr", "26")).To(MatchError(ContainSubstring("parse gh pr view output")))
			Expect(stdout.String()).To(BeEmpty())
		})

		DescribeTable("fails loud on gh api graphql output it cannot use",
			func(ctx SpecContext, pages, want string) {
				writeView(nil)
				Expect(os.WriteFile(pagesFile, []byte(pages), 0o600)).To(Succeed())

				Expect(run(ctx, "", "--pr", "26")).To(MatchError(ContainSubstring(want)))
				Expect(stdout.String()).To(BeEmpty())
			},
			Entry("not JSON", `{not json`, "parse gh api graphql output"),
			Entry("not slurped", `{"data":{}}`, "parse gh api graphql output"),
			Entry("no pages", `[]`, "no pages"),
			Entry("no repository", `[{"data":{"repository":null}}]`, "no pull request o/r#26"),
			Entry("no pull request", `[{"data":{"repository":{"pullRequest":null}}}]`, "no pull request o/r#26"),
			Entry("a commit without an oid",
				`[{"data":{"repository":{"pullRequest":{"headRefOid":"h2","commits":{"totalCount":1,"nodes":[{"commit":{"message":"x"}}]}}}}}]`,
				"a commit has no oid"),
			Entry("a page without a head",
				`[{"data":{"repository":{"pullRequest":{"commits":{"totalCount":0,"nodes":[]}}}}}]`,
				"no headRefOid"),
		)

		DescribeTable("fails loud on gh pr view JSON without a field it asked for",
			func(ctx SpecContext, set map[string]any, want string) {
				writeView(set)

				Expect(run(ctx, "", "--pr", "26")).To(MatchError(ContainSubstring(want)))
				Expect(stdout.String()).To(BeEmpty())
			},
			Entry("title missing", map[string]any{"title": nil}, "no title field"),
			Entry("title null", map[string]any{"title": jsonNull{}}, "no title field"),
			Entry("body missing", map[string]any{"body": nil}, "no body field"),
			Entry("body null", map[string]any{"body": jsonNull{}}, "no body field"),
			Entry("url missing", map[string]any{"url": nil}, "no url field"),
			Entry("headRefOid missing", map[string]any{"headRefOid": nil}, "no headRefOid field"),
			Entry("headRefOid empty", map[string]any{"headRefOid": ""}, "no headRefOid field"),
			Entry("headRefName missing", map[string]any{"headRefName": nil}, "no headRefName field"),
			Entry("headRepositoryOwner null", map[string]any{"headRepositoryOwner": jsonNull{}}, "no headRepositoryOwner field"),
			Entry("headRepositoryOwner without a login", map[string]any{"headRepositoryOwner": map[string]any{}}, "no headRepositoryOwner field"),
		)
	})

	It("fails loud when stdout cannot be written", func(ctx SpecContext) {
		err := closekw.Run(ctx, []string{"--description", "-"}, strings.NewReader("Fixes #1\n"), failingWriter{}, stderr)
		Expect(err).To(MatchError(ContainSubstring("write")))
	})
})

// jsonNull marshals as JSON null, where a nil map value drops the field.
type jsonNull struct{}

func (jsonNull) MarshalJSON() ([]byte, error) { return []byte("null"), nil }

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("disk full") }
