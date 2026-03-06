// Package git provides a thin Go wrapper around the git CLI.
// It covers status, diff, staging, committing, worktree management,
// repo discovery, and file watching with debounce.
package git

import (
	"strings"
)

// Repository provides high-level git operations for a single repo.
type Repository struct {
	cli *CLI
}

// NewRepository returns a Repository rooted at repoPath.
func NewRepository(repoPath string) *Repository {
	return &Repository{cli: NewCLI(repoPath)}
}

// Branch returns branch info including upstream tracking.
func (r *Repository) Branch() (BranchInfo, error) {
	name, err := r.cli.Run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return BranchInfo{}, err
	}

	info := BranchInfo{Name: name, IsDetached: name == "HEAD"}

	upstream, _ := r.cli.Run("rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	if upstream != "" && !strings.HasPrefix(upstream, "fatal") {
		info.Upstream = upstream
		ahead, _ := r.cli.Run("rev-list", "--count", upstream+"..HEAD")
		behind, _ := r.cli.Run("rev-list", "--count", "HEAD.."+upstream)
		info.Ahead = parseInt(ahead)
		info.Behind = parseInt(behind)
	}

	return info, nil
}

// Status returns the list of changed files.
func (r *Repository) Status() ([]FileStatus, error) {
	lines, err := r.cli.RunLines("status", "--porcelain=v1", "-u")
	if err != nil {
		return nil, err
	}

	var result []FileStatus
	for _, line := range lines {
		if len(line) < 4 {
			continue
		}
		fs := FileStatus{
			Staged:   FileStatusCode(string(line[0])),
			Unstaged: FileStatusCode(string(line[1])),
			Path:     strings.TrimSpace(line[3:]),
		}
		// Handle renames: "R old -> new"
		if fs.Staged == StatusRenamed || fs.Unstaged == StatusRenamed {
			parts := strings.SplitN(fs.Path, " -> ", 2)
			if len(parts) == 2 {
				fs.OldPath = parts[0]
				fs.Path = parts[1]
			}
		}
		result = append(result, fs)
	}
	return result, nil
}

// Diff returns the diff for a file. If staged is true, returns the staged diff.
func (r *Repository) Diff(filePath string, staged bool) ([]DiffLine, error) {
	args := []string{"diff"}
	if staged {
		args = append(args, "--cached")
	}
	args = append(args, "--", filePath)

	lines, err := r.cli.RunLines(args...)
	if err != nil {
		return nil, err
	}

	var result []DiffLine
	for _, line := range lines {
		dl := parseDiffLine(line)
		result = append(result, dl)
	}
	return result, nil
}

// Stage stages the given file.
func (r *Repository) Stage(filePath string) error {
	_, err := r.cli.Run("add", "--", filePath)
	return err
}

// StageAll stages all changes.
func (r *Repository) StageAll() error {
	_, err := r.cli.Run("add", "-A")
	return err
}

// Unstage unstages the given file.
func (r *Repository) Unstage(filePath string) error {
	_, err := r.cli.Run("restore", "--staged", "--", filePath)
	return err
}

// UnstageAll unstages everything.
func (r *Repository) UnstageAll() error {
	_, err := r.cli.Run("restore", "--staged", ".")
	return err
}

// Discard discards unstaged changes for a file.
func (r *Repository) Discard(filePath string) error {
	_, err := r.cli.Run("restore", "--", filePath)
	return err
}

// Commit creates a commit with the given message.
func (r *Repository) Commit(message string) error {
	_, err := r.cli.Run("commit", "-m", message)
	return err
}

// NumStat returns a compact diff summary (insertions, deletions, files).
func (r *Repository) NumStat() (RepoStats, error) {
	out, err := r.cli.Run("diff", "--numstat", "HEAD")
	if err != nil {
		return RepoStats{}, nil // not an error if HEAD doesn't exist yet
	}

	var stats RepoStats
	for _, line := range strings.Split(out, "\n") {
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		stats.Insertions += parseInt(parts[0])
		stats.Deletions += parseInt(parts[1])
		stats.Files++
	}
	return stats, nil
}

// LsFiles returns tracked files, suitable for the file finder.
func (r *Repository) LsFiles() ([]string, error) {
	return r.cli.RunLines("ls-files")
}

func parseDiffLine(line string) DiffLine {
	if strings.HasPrefix(line, "@@") {
		return DiffLine{Type: DiffLineHunkHeader, Content: line}
	}
	if strings.HasPrefix(line, "diff ") || strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
		return DiffLine{Type: DiffLineFileHeader, Content: line}
	}
	if strings.HasPrefix(line, "+") {
		return DiffLine{Type: DiffLineAdded, Content: line}
	}
	if strings.HasPrefix(line, "-") {
		return DiffLine{Type: DiffLineDeleted, Content: line}
	}
	return DiffLine{Type: DiffLineContext, Content: line}
}

func parseInt(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}
