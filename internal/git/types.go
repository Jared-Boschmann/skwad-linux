package git

// FileStatusCode matches standard git XY status codes.
type FileStatusCode string

const (
	StatusModified  FileStatusCode = "M"
	StatusAdded     FileStatusCode = "A"
	StatusDeleted   FileStatusCode = "D"
	StatusRenamed   FileStatusCode = "R"
	StatusCopied    FileStatusCode = "C"
	StatusUntracked FileStatusCode = "?"
	StatusIgnored   FileStatusCode = "!"
)

// FileStatus represents one file in a git status output.
type FileStatus struct {
	Path       string
	OldPath    string // for renames
	Staged     FileStatusCode
	Unstaged   FileStatusCode
}

// DiffLineType distinguishes context, addition, deletion, and hunk header lines.
type DiffLineType int

const (
	DiffLineContext  DiffLineType = iota
	DiffLineAdded
	DiffLineDeleted
	DiffLineHunkHeader
	DiffLineFileHeader
)

// DiffLine is one line in a diff output.
type DiffLine struct {
	Type    DiffLineType
	Content string
}

// BranchInfo holds the current branch and tracking state.
type BranchInfo struct {
	Name     string
	Upstream string
	Ahead    int
	Behind   int
	IsDetached bool
}

// Worktree represents one entry from `git worktree list`.
type Worktree struct {
	Path   string
	Branch string
	HEAD   string
	Bare   bool
}

// RepoStats is a compact summary for the sidebar display.
type RepoStats struct {
	Insertions int
	Deletions  int
	Files      int
}
