package terminal

import "testing"

func TestStripANSI(t *testing.T) {
	cases := []struct{ in, want string }{
		{"\x1b[31mred\x1b[0m", "red"},
		{"\x1b[1;32mbold green\x1b[0m", "bold green"},
		{"plain text", "plain text"},
		{"\x1b]0;my title\x07", ""},
		{"no escapes", "no escapes"},
	}
	for _, tc := range cases {
		if got := StripANSI(tc.in); got != tc.want {
			t.Errorf("StripANSI(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestCleanTitle_StripsSpinners(t *testing.T) {
	cases := []struct{ in, want string }{
		{"⠋ claude working", "claude working"},
		{"● running", "running"},
		{"… thinking", "thinking"},
		{"normal title", "normal title"},
		{"  spaces  ", "spaces"},
	}
	for _, tc := range cases {
		if got := CleanTitle(tc.in); got != tc.want {
			t.Errorf("CleanTitle(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestCleanTitle_StripsStatusPrefixes(t *testing.T) {
	cases := []struct{ in, want string }{
		{"[running] my agent", "my agent"},
		{"[idle] my agent", "my agent"},
		{"(running) task", "task"},
	}
	for _, tc := range cases {
		if got := CleanTitle(tc.in); got != tc.want {
			t.Errorf("CleanTitle(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestExtractTitle_OSC0(t *testing.T) {
	// ESC ] 0 ; title BEL
	data := []byte("\x1b]0;my title\x07")
	title, ok := extractTitle(data)
	if !ok {
		t.Fatal("expected to extract title from OSC 0 sequence")
	}
	if title != "my title" {
		t.Errorf("got %q, want %q", title, "my title")
	}
}

func TestExtractTitle_OSC2(t *testing.T) {
	data := []byte("\x1b]2;window title\x07")
	title, ok := extractTitle(data)
	if !ok {
		t.Fatal("expected to extract title from OSC 2 sequence")
	}
	if title != "window title" {
		t.Errorf("got %q, want %q", title, "window title")
	}
}

func TestExtractTitle_NoSequence(t *testing.T) {
	_, ok := extractTitle([]byte("hello world"))
	if ok {
		t.Error("should not extract title from plain text")
	}
}

func TestExtractTitle_EmbeddedInOutput(t *testing.T) {
	// Title sequence surrounded by other output
	data := []byte("some output\x1b]0;agent title\x07more output")
	title, ok := extractTitle(data)
	if !ok {
		t.Fatal("expected to find title in mixed output")
	}
	if title != "agent title" {
		t.Errorf("got %q, want %q", title, "agent title")
	}
}
