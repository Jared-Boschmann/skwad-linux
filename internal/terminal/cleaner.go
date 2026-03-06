package terminal

import (
	"regexp"
	"strings"
)

// ansiEscape matches ANSI/VT100 escape sequences.
// OSC (] ... BEL or ] ... ST) must come before the generic C1 catch-all
// because ']' (0x5D) falls inside the [\x5C-\x5F] range.
var ansiEscape = regexp.MustCompile(`\x1b(?:\][^\x07\x1b]*(?:\x07|\x1b\\)|\[[0-?]*[ -/]*[@-~]|[@-Z\\-_])`)

// StripANSI removes ANSI escape sequences from s.
func StripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

// spinnerChars are leading characters commonly emitted by AI CLIs to indicate
// activity. They are stripped from terminal titles before display.
var spinnerChars = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏", "◐", "◓", "◑", "◒", "⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷", "⠁", "⠂", "⠄", "⡀", "⢀", "⠠", "⠐", "⠈", "✓", "✗", "●", "○", "►", "▶", "…", "·"}

// CleanTitle strips ANSI sequences and leading spinner characters from a
// terminal title string, returning a clean display name.
func CleanTitle(title string) string {
	title = StripANSI(title)
	title = strings.TrimSpace(title)

	for _, ch := range spinnerChars {
		title = strings.TrimPrefix(title, ch)
		title = strings.TrimSpace(title)
	}

	// Strip common status prefixes like "[running] " or "(idle) "
	for _, prefix := range []string{"[running]", "[idle]", "[input]", "[error]", "(running)", "(idle)"} {
		if strings.HasPrefix(strings.ToLower(title), prefix) {
			title = strings.TrimSpace(title[len(prefix):])
		}
	}

	return title
}
