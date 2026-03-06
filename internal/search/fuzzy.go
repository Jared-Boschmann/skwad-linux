// Package search provides a lightweight fuzzy file-path scorer.
// It ranks results by subsequence match quality, with bonuses for
// consecutive characters and matches immediately after path separators.
package search

import "strings"

// Result is a file path with its fuzzy match score and match indices.
type Result struct {
	Path    string
	Score   int
	Indices []int // character positions that matched query runes
}

// FuzzySearch filters and ranks paths by query, returning up to maxResults.
func FuzzySearch(paths []string, query string, maxResults int) []Result {
	query = strings.ToLower(query)
	var results []Result

	for _, path := range paths {
		score, indices := score(strings.ToLower(path), query)
		if score > 0 {
			results = append(results, Result{Path: path, Score: score, Indices: indices})
		}
	}

	// Sort descending by score.
	sortResults(results)

	if len(results) > maxResults {
		return results[:maxResults]
	}
	return results
}

// score returns a positive score and match indices if query is a subsequence of
// target, 0 otherwise. Consecutive matches and matches near the end of path
// components score higher.
func score(target, query string) (int, []int) {
	if query == "" {
		return 0, nil
	}

	tr := []rune(target)
	qr := []rune(query)

	indices := make([]int, 0, len(qr))
	ti := 0
	for qi := 0; qi < len(qr); qi++ {
		found := false
		for ; ti < len(tr); ti++ {
			if tr[ti] == qr[qi] {
				indices = append(indices, ti)
				ti++
				found = true
				break
			}
		}
		if !found {
			return 0, nil
		}
	}

	// Base score from total matched characters.
	sc := len(qr) * 10

	// Bonus for consecutive matches.
	for i := 1; i < len(indices); i++ {
		if indices[i] == indices[i-1]+1 {
			sc += 5
		}
	}

	// Bonus for matches after path separators.
	for _, idx := range indices {
		if idx == 0 || tr[idx-1] == '/' {
			sc += 8
		}
	}

	// Penalty for total target length (prefer shorter matches).
	sc -= len(tr) / 4

	return sc, indices
}

func sortResults(results []Result) {
	// Simple insertion sort — result sets are small (<50k before filtering).
	for i := 1; i < len(results); i++ {
		key := results[i]
		j := i - 1
		for j >= 0 && results[j].Score < key.Score {
			results[j+1] = results[j]
			j--
		}
		results[j+1] = key
	}
}
