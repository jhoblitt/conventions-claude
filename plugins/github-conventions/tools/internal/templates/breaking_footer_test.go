package templates_test

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The template ships inside the skill, outside every module, so the spec finds
// it by walking up from this file; CLAUDE_PLUGIN_ROOT is not set under go test.
const breakingFooterTemplate = "../../../skills/github-conventions/templates/breaking-footer/main.go"

type gitRepo struct {
	dir string
}

type runResult struct {
	code   int
	stdout string
	stderr string
}

func newGitRepo(ctx SpecContext) *gitRepo {
	GinkgoHelper()

	repo := &gitRepo{dir: GinkgoT().TempDir()}
	repo.git(ctx, "init", "--initial-branch=main")
	repo.git(ctx, "config", "user.name", "Spec Runner")
	repo.git(ctx, "config", "user.email", "spec@example.invalid")
	repo.git(ctx, "config", "commit.gpgsign", "false")

	return repo
}

func (r *gitRepo) git(ctx SpecContext, args ...string) string {
	GinkgoHelper()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = r.dir
	// The developer's own git config must not reach the fixture: a global
	// hooksPath, template dir, or signing key would change what commits.
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")

	out, err := cmd.CombinedOutput()
	Expect(err).NotTo(HaveOccurred(), "git %s: %s", strings.Join(args, " "), out)

	return strings.TrimSpace(string(out))
}

func (r *gitRepo) commit(ctx SpecContext, subject, body string) string {
	GinkgoHelper()

	args := []string{"commit", "--allow-empty", "-m", subject}
	if body != "" {
		args = append(args, "-m", body)
	}
	r.git(ctx, args...)

	return r.git(ctx, "rev-parse", "HEAD")
}

func (r *gitRepo) checkBreakingFooters(ctx SpecContext, base, head string) runResult {
	GinkgoHelper()

	template, err := filepath.Abs(breakingFooterTemplate)
	Expect(err).NotTo(HaveOccurred())

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "go", "run", template, base, head)
	cmd.Dir = r.dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	code := 0
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		Expect(errors.As(err, &exitErr)).To(BeTrue(), "go run: %v: %s", err, stderr.String())
		code = exitErr.ExitCode()
	}

	return runResult{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

var _ = Describe("the breaking-footer template", func() {
	It("passes a range whose commits declare no footer", func(ctx SpecContext) {
		repo := newGitRepo(ctx)
		base := repo.commit(ctx, "chore: root", "")
		repo.commit(ctx, "feat: add a thing", "Nothing breaking here.")
		head := repo.commit(ctx, "fix: repair a thing", "")

		result := repo.checkBreakingFooters(ctx, base, head)

		Expect(result.code).To(Equal(0), "stdout: %s\nstderr: %s", result.stdout, result.stderr)
		Expect(result.stdout).To(ContainSubstring("checked 2 commit(s): no undeclared breaking-change footers"))
	})

	It("fails a body footer the subject never declared, naming the commit and the lines", func(ctx SpecContext) {
		repo := newGitRepo(ctx)
		base := repo.commit(ctx, "chore: root", "")
		head := repo.commit(ctx, "feat: add a thing",
			"Some prose.\n\nBREAKING CHANGE: the api moved\n  * BREAKING-CHANGE the second spelling")

		result := repo.checkBreakingFooters(ctx, base, head)

		Expect(result.code).To(Equal(1), "stdout: %s\nstderr: %s", result.stdout, result.stderr)
		Expect(result.stdout).To(ContainSubstring(head[:8] + " declares no break in its subject"))
		Expect(result.stdout).To(ContainSubstring("  subject: feat: add a thing"))
		Expect(result.stdout).To(ContainSubstring("  body line 3:BREAKING CHANGE: the api moved"))
		Expect(result.stdout).To(ContainSubstring("  body line 4:  * BREAKING-CHANGE the second spelling"))
		Expect(result.stderr).To(ContainSubstring("Either declare the break in the subject"))
	})

	It("passes a body footer whose subject carries the bang", func(ctx SpecContext) {
		repo := newGitRepo(ctx)
		base := repo.commit(ctx, "chore: root", "")
		repo.commit(ctx, "feat!: add a thing", "BREAKING CHANGE: the api moved")
		head := repo.commit(ctx, "fix(api)!: repair a thing", "BREAKING-CHANGE: the api moved again")

		result := repo.checkBreakingFooters(ctx, base, head)

		Expect(result.code).To(Equal(0), "stdout: %s\nstderr: %s", result.stdout, result.stderr)
		Expect(result.stdout).To(ContainSubstring("checked 2 commit(s)"))
	})

	It("ignores a footer carried by a merge commit", func(ctx SpecContext) {
		repo := newGitRepo(ctx)
		base := repo.commit(ctx, "chore: root", "")
		repo.git(ctx, "switch", "--create", "topic")
		repo.commit(ctx, "feat: topic work", "")
		repo.git(ctx, "switch", "main")
		repo.commit(ctx, "fix: main work", "")
		repo.git(ctx, "merge", "--no-ff", "-m", "chore: merge topic", "-m", "BREAKING CHANGE: from the merge", "topic")
		head := repo.git(ctx, "rev-parse", "HEAD")

		result := repo.checkBreakingFooters(ctx, base, head)

		Expect(result.code).To(Equal(0), "stdout: %s\nstderr: %s", result.stdout, result.stderr)
		Expect(result.stdout).To(ContainSubstring("checked 2 commit(s)"))
	})

	It("passes a body that mentions the keyword mid-line", func(ctx SpecContext) {
		repo := newGitRepo(ctx)
		base := repo.commit(ctx, "chore: root", "")
		head := repo.commit(ctx, "feat: add a thing",
			"This one is not a BREAKING CHANGE: it only talks about them.")

		result := repo.checkBreakingFooters(ctx, base, head)

		Expect(result.code).To(Equal(0), "stdout: %s\nstderr: %s", result.stdout, result.stderr)
		Expect(result.stdout).To(ContainSubstring("checked 1 commit(s)"))
	})
})
