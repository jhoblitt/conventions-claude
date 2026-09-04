package audit_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jhoblitt/conventions-claude/plugins/github-conventions/tools/internal/audit"
)

var _ = Describe("ParseSlug", func() {
	DescribeTable("reads owner/repo from an origin URL",
		func(remote, want string) {
			got, err := audit.ParseSlug(remote)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(want))
		},
		Entry("https", "https://github.com/jhoblitt/conventions-claude", "jhoblitt/conventions-claude"),
		Entry("https with .git", "https://github.com/jhoblitt/conventions-claude.git", "jhoblitt/conventions-claude"),
		Entry("ssh", "git@github.com:jhoblitt/conventions-claude.git", "jhoblitt/conventions-claude"),
		Entry("trailing slash", "https://github.com/jhoblitt/conventions-claude/", "jhoblitt/conventions-claude"),
	)

	It("fails on a URL with no repository", func() {
		_, err := audit.ParseSlug("github.com")
		Expect(err).To(HaveOccurred())
	})
})
