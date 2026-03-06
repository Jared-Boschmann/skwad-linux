package search

import (
	"testing"
)

func TestFuzzySearch_BasicMatch(t *testing.T) {
	paths := []string{"internal/agent/manager.go", "internal/mcp/server.go", "cmd/skwad/main.go"}
	results := FuzzySearch(paths, "main", 10)
	if len(results) == 0 {
		t.Fatal("expected at least one result for 'main'")
	}
	if results[0].Path != "cmd/skwad/main.go" {
		t.Errorf("expected cmd/skwad/main.go first, got %s", results[0].Path)
	}
}

func TestFuzzySearch_NoMatch(t *testing.T) {
	paths := []string{"foo.go", "bar.go"}
	results := FuzzySearch(paths, "zzz", 10)
	if len(results) != 0 {
		t.Errorf("expected no results, got %d", len(results))
	}
}

func TestFuzzySearch_Subsequence(t *testing.T) {
	paths := []string{"internal/agent/manager.go"}
	results := FuzzySearch(paths, "agmgr", 10)
	if len(results) == 0 {
		t.Fatal("expected subsequence match for 'agmgr'")
	}
}

func TestFuzzySearch_RankedByScore(t *testing.T) {
	// "server.go" should score higher than "settings_window.go" for query "server"
	paths := []string{"internal/ui/settings_window.go", "internal/mcp/server.go"}
	results := FuzzySearch(paths, "server", 10)
	if len(results) < 2 {
		t.Skip("not enough results to compare ranking")
	}
	if results[0].Path != "internal/mcp/server.go" {
		t.Errorf("expected server.go to rank first, got %s", results[0].Path)
	}
}

func TestFuzzySearch_MaxResults(t *testing.T) {
	paths := make([]string, 20)
	for i := range paths {
		paths[i] = "file.go"
	}
	results := FuzzySearch(paths, "file", 5)
	if len(results) > 5 {
		t.Errorf("expected at most 5 results, got %d", len(results))
	}
}

func TestFuzzySearch_EmptyQuery(t *testing.T) {
	paths := []string{"foo.go"}
	results := FuzzySearch(paths, "", 10)
	if len(results) != 0 {
		t.Errorf("empty query should return no results, got %d", len(results))
	}
}

func TestFuzzySearch_EmptyPaths(t *testing.T) {
	results := FuzzySearch(nil, "foo", 10)
	if len(results) != 0 {
		t.Errorf("nil paths should return no results, got %d", len(results))
	}
}

func TestScore_ConsecutiveBonus(t *testing.T) {
	// "abcd" matching consecutively should score higher than spread-out match
	scoreConsec, _ := score("abcd", "abc")
	scoreSpread, _ := score("axbxcx", "abc")
	if scoreConsec <= scoreSpread {
		t.Errorf("consecutive match (%d) should outscore spread match (%d)", scoreConsec, scoreSpread)
	}
}
