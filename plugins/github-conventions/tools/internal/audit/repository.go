package audit

import (
	"context"
	"log/slog"
	"strconv"
)

// Repository is the part of a GitHub repository's settings this audit reads.
type Repository struct {
	FullName            string `json:"full_name"`
	DeleteBranchOnMerge bool   `json:"delete_branch_on_merge"`
}

func repositoryRow(ctx context.Context, opts Options) Row {
	const canon = "delete_branch_on_merge is on: a merged PR's branch is deleted"

	if !opts.Remote {
		return skippedRow("repository", "delete-branch-on-merge", canon)
	}

	repo, err := opts.remoteLookup().Repository(ctx, opts.Dir)
	if err != nil {
		opts.Logger.WarnContext(ctx, "repository lookup failed", slog.Any("error", err))

		return gapRow("repository", "delete-branch-on-merge", "gh api failed: "+err.Error(), canon, repoEditFix(""))
	}

	current := "delete_branch_on_merge: " + strconv.FormatBool(repo.DeleteBranchOnMerge)
	if repo.DeleteBranchOnMerge {
		return okRow("repository", "delete-branch-on-merge", current, canon)
	}

	return gapRow("repository", "delete-branch-on-merge", current, canon, repoEditFix(repo.FullName))
}

// repoEditFix names the repository the API reported, so the command the user
// approves carries its target instead of leaving it to gh's base-repository
// resolution, which in a clone with several remotes may pick another. With no
// name (the lookup failed), gh reads it from the checkout's remote.
func repoEditFix(fullName string) string {
	if fullName == "" {
		return "gh repo edit --delete-branch-on-merge"
	}

	return "gh repo edit " + fullName + " --delete-branch-on-merge"
}

// Repository implements [RemoteLookup] against the GitHub API.
func (l ghRemoteLookup) Repository(ctx context.Context, dir string) (Repository, error) {
	slug, err := originSlug(ctx, dir)
	if err != nil {
		return Repository{}, err
	}

	var repo Repository
	if err := l.api(ctx, dir, &repo, "repos/"+slug); err != nil {
		return Repository{}, err
	}

	return repo, nil
}
