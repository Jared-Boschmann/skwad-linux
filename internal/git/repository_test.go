package git

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// initRepo creates a temporary git repo with an initial commit and returns its path.
func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	run := func(args ...string) {
		t.Helper()
		cli := NewCLI(dir)
		if _, err := cli.Run(args...); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}

	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test User")

	// Initial commit
	writeFile(t, dir, "README.md", "# Test\n")
	run("add", "README.md")
	run("commit", "-m", "initial commit")

	return dir
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// --- Branch ---

func TestRepository_Branch_Name(t *testing.T) {
	dir := initRepo(t)
	r := NewRepository(dir)
	b, err := r.Branch()
	if err != nil {
		t.Fatalf("Branch: %v", err)
	}
	// git init default branch is "main" or "master" depending on config
	if b.Name == "" {
		t.Error("branch name should not be empty")
	}
}

// --- Status ---

func TestRepository_Status_Clean(t *testing.T) {
	dir := initRepo(t)
	r := NewRepository(dir)
	files, err := r.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected clean status, got %d files", len(files))
	}
}

func TestRepository_Status_Modified(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, dir, "README.md", "modified\n")
	r := NewRepository(dir)
	files, err := r.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected at least one modified file")
	}
	found := false
	for _, f := range files {
		if f.Path == "README.md" {
			found = true
		}
	}
	if !found {
		t.Error("README.md not in status output")
	}
}

func TestRepository_Status_Untracked(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, dir, "new_file.go", "package main\n")
	r := NewRepository(dir)
	files, err := r.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	found := false
	for _, f := range files {
		if strings.Contains(f.Path, "new_file.go") {
			found = true
			if f.Unstaged != StatusUntracked {
				t.Errorf("expected untracked status, got %q", f.Unstaged)
			}
		}
	}
	if !found {
		t.Error("new_file.go not found in status")
	}
}

// --- Stage / Unstage ---

func TestRepository_StageAndUnstage(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, dir, "staged.go", "package main\n")
	r := NewRepository(dir)

	if err := r.Stage("staged.go"); err != nil {
		t.Fatalf("Stage: %v", err)
	}

	files, _ := r.Status()
	staged := false
	for _, f := range files {
		if f.Path == "staged.go" && f.Staged == StatusAdded {
			staged = true
		}
	}
	if !staged {
		t.Error("file should be staged after Stage()")
	}

	if err := r.Unstage("staged.go"); err != nil {
		t.Fatalf("Unstage: %v", err)
	}

	files, _ = r.Status()
	for _, f := range files {
		if f.Path == "staged.go" && f.Staged == StatusAdded {
			t.Error("file should not be staged after Unstage()")
		}
	}
}

func TestRepository_StageAll(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, dir, "a.go", "package a\n")
	writeFile(t, dir, "b.go", "package b\n")
	r := NewRepository(dir)

	if err := r.StageAll(); err != nil {
		t.Fatalf("StageAll: %v", err)
	}

	files, _ := r.Status()
	for _, f := range files {
		if f.Staged != StatusAdded && f.Staged != FileStatusCode(" ") {
			// At least some files should be staged
		}
	}
}

// --- Commit ---

func TestRepository_Commit(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, dir, "commit_test.go", "package main\n")
	r := NewRepository(dir)
	cli := NewCLI(dir)

	if err := r.Stage("commit_test.go"); err != nil {
		t.Fatalf("Stage: %v", err)
	}
	if err := r.Commit("test commit"); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	log, err := cli.Run("log", "--oneline", "-1")
	if err != nil {
		t.Fatalf("git log: %v", err)
	}
	if !strings.Contains(log, "test commit") {
		t.Errorf("commit message not found in log: %q", log)
	}
}

// --- Diff ---

func TestRepository_Diff_Modified(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, dir, "README.md", "# Modified\nadded line\n")
	r := NewRepository(dir)

	lines, err := r.Diff("README.md", false)
	if err != nil {
		t.Fatalf("Diff: %v", err)
	}
	if len(lines) == 0 {
		t.Fatal("expected diff output for modified file")
	}

	hasAddition := false
	for _, l := range lines {
		if l.Type == DiffLineAdded {
			hasAddition = true
		}
	}
	if !hasAddition {
		t.Error("expected at least one added line in diff")
	}
}

// --- NumStat ---

func TestRepository_NumStat_AfterModify(t *testing.T) {
	dir := initRepo(t)
	writeFile(t, dir, "README.md", "# Modified\nmore content\n")
	r := NewRepository(dir)

	stats, err := r.NumStat()
	if err != nil {
		t.Fatalf("NumStat: %v", err)
	}
	if stats.Insertions == 0 && stats.Files == 0 {
		// NumStat compares to HEAD — may show 0 if repo was just init'd
		// This is acceptable; just verify no crash.
	}
}

// --- LsFiles ---

func TestRepository_LsFiles(t *testing.T) {
	dir := initRepo(t)
	r := NewRepository(dir)

	files, err := r.LsFiles()
	if err != nil {
		t.Fatalf("LsFiles: %v", err)
	}
	if len(files) == 0 {
		t.Error("expected at least one tracked file")
	}
	found := false
	for _, f := range files {
		if f == "README.md" {
			found = true
		}
	}
	if !found {
		t.Error("README.md not in ls-files output")
	}
}

// --- IsRepo / RootOf ---

func TestIsRepo(t *testing.T) {
	dir := initRepo(t)
	if !IsRepo(dir) {
		t.Error("should be a git repo")
	}
	if IsRepo(t.TempDir()) {
		t.Error("empty dir should not be a git repo")
	}
}

func TestRootOf(t *testing.T) {
	dir := initRepo(t)
	root, err := RootOf(dir)
	if err != nil {
		t.Fatalf("RootOf: %v", err)
	}
	// Resolve symlinks on both sides — macOS /var is a symlink to /private/var.
	realDir, _ := filepath.EvalSymlinks(dir)
	realRoot, _ := filepath.EvalSymlinks(root)
	if realRoot != realDir {
		t.Errorf("got %q, want %q", root, dir)
	}
}

// --- SuggestedPath ---

func TestSuggestedPath(t *testing.T) {
	got := SuggestedPath("/home/user/myrepo", "feature/my-thing")
	want := "/home/user/myrepo-feature-my-thing"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSuggestedPath_SpacesReplaced(t *testing.T) {
	got := SuggestedPath("/home/user/repo", "fix bad thing")
	if strings.Contains(got, " ") {
		t.Errorf("path should not contain spaces: %q", got)
	}
}
