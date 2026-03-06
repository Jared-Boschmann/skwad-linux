package git

import (
	"os"
	"path/filepath"
	"strings"
)

// ExcludedDirs is the set of directory names skipped during file indexing.
// Matches REQUIREMENTS §3.12.
var ExcludedDirs = map[string]bool{
	".git":        true,
	"node_modules": true,
	".build":      true,
	"__pycache__": true,
	".DS_Store":   true,
	".svn":        true,
	".hg":         true,
	"Pods":        true,
	"DerivedData": true,
	".idea":       true,
	".vscode":     true,
	"vendor":      true,
	"dist":        true,
	"build":       true,
}

// IsExcluded reports whether a path component should be skipped in file search.
func IsExcluded(name string) bool {
	return ExcludedDirs[name] || strings.HasPrefix(name, ".")
}

// CommonSourceDirs is the ordered list of directories checked for a source base folder.
var CommonSourceDirs = []string{
	"src", "dev", "code", "projects", "repos", "source", "sources",
	"workspace", "workspaces", "git", "github", "Development", "work", "coding",
}

// AutoDetectSourceDir returns the first common source directory that exists
// under the user's home directory.
func AutoDetectSourceDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	for _, dir := range CommonSourceDirs {
		path := filepath.Join(home, dir)
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return path
		}
	}
	return ""
}

// DiscoverRepos scans baseDir (non-recursively into sub-dirs one level deep)
// for git repositories and returns their root paths.
func DiscoverRepos(baseDir string) ([]string, error) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return nil, err
	}

	var repos []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := filepath.Join(baseDir, entry.Name())
		if IsRepo(path) {
			repos = append(repos, path)
		}
	}
	return repos, nil
}
