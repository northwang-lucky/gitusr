package gitcmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// IsGitRepo checks whether the current working directory is inside a git
// repository by verifying that a .git directory or file exists.
func IsGitRepo() bool {
	_, err := os.Stat(".git")
	return err == nil
}

// GetConfig retrieves a git config value for the given user key.
// If global is true, the --global flag is passed to git config.
// The returned string is trimmed of leading and trailing whitespace.
func GetConfig(key string, global bool) (string, error) {
	args := []string{"config"}
	if global {
		args = append(args, "--global")
	}
	args = append(args, "--get", "user."+key)

	cmd := exec.Command("git", args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git config failed: %s", strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

// SetConfig sets a git config value for the given user key.
// If global is true, the --global flag is passed to git config.
func SetConfig(key string, value string, global bool) error {
	args := []string{"config"}
	if global {
		args = append(args, "--global")
	}
	args = append(args, "user."+key, value)

	cmd := exec.Command("git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git config failed: %s", strings.TrimSpace(stderr.String()))
	}
	return nil
}

// CreateBackupBranch creates a local git branch with the given name.
func CreateBackupBranch(name string) error {
	cmd := exec.Command("git", "branch", name)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git branch failed: %s", strings.TrimSpace(stderr.String()))
	}
	return nil
}

// filterRepoCallback returns a Python snippet suitable for git-filter-repo's
// --commit-callback that replaces author and committer information for commits
// whose author email matches oldEmail.
func filterRepoCallback(oldEmail, newName, newEmail string) string {
	return fmt.Sprintf(`
if commit.author_email == b'%s':
    commit.author_name = b'%s'
    commit.author_email = b'%s'
    commit.committer_name = b'%s'
    commit.committer_email = b'%s'
`, oldEmail, newName, newEmail, newName, newEmail)
}

// FilterRepo rewrites git history using git-filter-repo, replacing the author
// and committer information for commits matching oldEmail with newName and newEmail.
// It first verifies that git-filter-repo is installed; if not, an error with
// installation instructions is returned.
func FilterRepo(oldEmail string, newName string, newEmail string) error {
	// Verify git-filter-repo is installed
	checkCmd := exec.Command("git", "filter-repo", "--help")
	if err := checkCmd.Run(); err != nil {
		return fmt.Errorf(
			"git-filter-repo is not installed. Install it with: pip install git-filter-repo",
		)
	}

	callback := filterRepoCallback(oldEmail, newName, newEmail)
	cmd := exec.Command("git", "filter-repo", "--commit-callback", callback, "--force")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git filter-repo failed: %s", strings.TrimSpace(stderr.String()))
	}
	return nil
}

// HasUncommittedChanges returns true if the working tree has uncommitted
// changes (tracked modified files or untracked files), determined by
// running git status --porcelain.
func HasUncommittedChanges() bool {
	cmd := exec.Command("git", "status", "--porcelain")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return false
	}
	return strings.TrimSpace(stdout.String()) != ""
}
