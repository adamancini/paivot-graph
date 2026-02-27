package lifecycle

import (
	"testing"
)

func TestContainsTriggerPhrase_Matches(t *testing.T) {
	cases := []struct {
		prompt string
		want   bool
	}{
		{"I want to use Paivot for this project", true},
		{"Paivot this feature request", true},
		{"Let's run Paivot on the requirements", true},
		{"Please engage Paivot now", true},
		{"Build the backend with Paivot", true},
		{"USE PAIVOT", true},
		{"just fix the bug", false},
		{"", false},
		{"paivot is cool but I didn't say the magic words", false},
		{"let me use paivot to help", true},
	}

	for _, tc := range cases {
		t.Run(tc.prompt, func(t *testing.T) {
			got := containsTriggerPhrase(tc.prompt)
			if got != tc.want {
				t.Errorf("containsTriggerPhrase(%q) = %v, want %v", tc.prompt, got, tc.want)
			}
		})
	}
}

func TestContainsTriggerPhrase_CaseInsensitive(t *testing.T) {
	if !containsTriggerPhrase("USE PAIVOT") {
		t.Error("expected case-insensitive match for USE PAIVOT")
	}
	if !containsTriggerPhrase("Use Paivot") {
		t.Error("expected case-insensitive match for Use Paivot")
	}
	if !containsTriggerPhrase("use paivot") {
		t.Error("expected case-insensitive match for use paivot")
	}
}
