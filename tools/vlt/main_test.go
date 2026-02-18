package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCmd    string
		wantParams map[string]string
		wantFlags  map[string]bool
	}{
		{
			name:       "read command",
			args:       []string{"vault=Claude", "read", "file=Session Operating Mode"},
			wantCmd:    "read",
			wantParams: map[string]string{"vault": "Claude", "file": "Session Operating Mode"},
			wantFlags:  map[string]bool{},
		},
		{
			name:       "create with silent flag",
			args:       []string{"vault=Claude", "create", "name=My Note", "path=_inbox/My Note.md", "content=# Hello", "silent"},
			wantCmd:    "create",
			wantParams: map[string]string{"vault": "Claude", "name": "My Note", "path": "_inbox/My Note.md", "content": "# Hello"},
			wantFlags:  map[string]bool{"silent": true},
		},
		{
			name:       "search command",
			args:       []string{"vault=Claude", "search", "query=paivot"},
			wantCmd:    "search",
			wantParams: map[string]string{"vault": "Claude", "query": "paivot"},
			wantFlags:  map[string]bool{},
		},
		{
			name:       "move command",
			args:       []string{"vault=Claude", "move", "path=_inbox/Note.md", "to=decisions/Note.md"},
			wantCmd:    "move",
			wantParams: map[string]string{"vault": "Claude", "path": "_inbox/Note.md", "to": "decisions/Note.md"},
			wantFlags:  map[string]bool{},
		},
		{
			name:       "property:set command",
			args:       []string{"vault=Claude", "property:set", "file=Note", "name=status", "value=archived"},
			wantCmd:    "property:set",
			wantParams: map[string]string{"vault": "Claude", "file": "Note", "name": "status", "value": "archived"},
			wantFlags:  map[string]bool{},
		},
		{
			name:       "content with equals sign",
			args:       []string{"vault=Claude", "create", "name=Note", "path=_inbox/Note.md", "content=key=value"},
			wantCmd:    "create",
			wantParams: map[string]string{"vault": "Claude", "name": "Note", "path": "_inbox/Note.md", "content": "key=value"},
			wantFlags:  map[string]bool{},
		},
		{
			name:       "quoted value stripping",
			args:       []string{`vault="Claude"`, "read", `file="My Note"`},
			wantCmd:    "read",
			wantParams: map[string]string{"vault": "Claude", "file": "My Note"},
			wantFlags:  map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, params, flags := parseArgs(tt.args)

			if cmd != tt.wantCmd {
				t.Errorf("cmd = %q, want %q", cmd, tt.wantCmd)
			}

			for k, want := range tt.wantParams {
				got, ok := params[k]
				if !ok {
					t.Errorf("missing param %q", k)
				} else if got != want {
					t.Errorf("param[%q] = %q, want %q", k, got, want)
				}
			}
			if len(params) != len(tt.wantParams) {
				t.Errorf("got %d params, want %d", len(params), len(tt.wantParams))
			}

			for k := range tt.wantFlags {
				if !flags[k] {
					t.Errorf("missing flag %q", k)
				}
			}
			if len(flags) != len(tt.wantFlags) {
				t.Errorf("got %d flags, want %d", len(flags), len(tt.wantFlags))
			}
		})
	}
}

func TestResolveNote(t *testing.T) {
	// Create a temporary vault
	vaultDir := t.TempDir()

	// Create directory structure
	os.MkdirAll(filepath.Join(vaultDir, "methodology"), 0755)
	os.MkdirAll(filepath.Join(vaultDir, "conventions"), 0755)
	os.MkdirAll(filepath.Join(vaultDir, "_inbox"), 0755)
	os.MkdirAll(filepath.Join(vaultDir, ".obsidian"), 0755)

	// Create test notes
	os.WriteFile(filepath.Join(vaultDir, "methodology", "Sr PM Agent.md"), []byte("# Sr PM"), 0644)
	os.WriteFile(filepath.Join(vaultDir, "conventions", "Session Operating Mode.md"), []byte("# SOM"), 0644)
	os.WriteFile(filepath.Join(vaultDir, ".obsidian", "hidden.md"), []byte("# Hidden"), 0644)

	tests := []struct {
		title   string
		wantRel string
		wantErr bool
	}{
		{"Sr PM Agent", "methodology/Sr PM Agent.md", false},
		{"Session Operating Mode", "conventions/Session Operating Mode.md", false},
		{"Nonexistent Note", "", true},
		{"hidden", "", true}, // should not find notes in .obsidian
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			path, err := resolveNote(vaultDir, tt.title)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got path %q", path)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			relPath, _ := filepath.Rel(vaultDir, path)
			if relPath != tt.wantRel {
				t.Errorf("got %q, want %q", relPath, tt.wantRel)
			}
		})
	}
}

func TestCmdCreateAndRead(t *testing.T) {
	vaultDir := t.TempDir()

	// Create a note
	params := map[string]string{
		"name":    "Test Note",
		"path":    "_inbox/Test Note.md",
		"content": "---\ntype: test\n---\n\n# Test Note\n\nHello world.\n",
	}
	if err := cmdCreate(vaultDir, params, false); err != nil {
		t.Fatalf("create: %v", err)
	}

	// Verify file exists
	fullPath := filepath.Join(vaultDir, "_inbox", "Test Note.md")
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
	if string(data) != params["content"] {
		t.Errorf("content mismatch:\ngot:  %q\nwant: %q", string(data), params["content"])
	}

	// Create again (should be a no-op, not overwrite)
	params["content"] = "overwritten"
	if err := cmdCreate(vaultDir, params, true); err != nil {
		t.Fatalf("create (duplicate): %v", err)
	}
	data, _ = os.ReadFile(fullPath)
	if string(data) == "overwritten" {
		t.Error("create overwrote existing note")
	}
}

func TestCmdAppend(t *testing.T) {
	vaultDir := t.TempDir()

	// Create a note to append to
	notePath := filepath.Join(vaultDir, "Test Append.md")
	os.WriteFile(notePath, []byte("# Test\n"), 0644)

	params := map[string]string{
		"file":    "Test Append",
		"content": "\n## Added section\n",
	}
	if err := cmdAppend(vaultDir, params); err != nil {
		t.Fatalf("append: %v", err)
	}

	data, _ := os.ReadFile(notePath)
	want := "# Test\n\n## Added section\n"
	if string(data) != want {
		t.Errorf("got %q, want %q", string(data), want)
	}
}

func TestCmdMove(t *testing.T) {
	vaultDir := t.TempDir()

	// Create source
	os.MkdirAll(filepath.Join(vaultDir, "_inbox"), 0755)
	srcPath := filepath.Join(vaultDir, "_inbox", "Note.md")
	os.WriteFile(srcPath, []byte("# Note"), 0644)

	params := map[string]string{
		"path": "_inbox/Note.md",
		"to":   "decisions/Note.md",
	}
	if err := cmdMove(vaultDir, params); err != nil {
		t.Fatalf("move: %v", err)
	}

	// Source should be gone
	if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
		t.Error("source file still exists after move")
	}

	// Destination should exist
	dstPath := filepath.Join(vaultDir, "decisions", "Note.md")
	data, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("destination not found: %v", err)
	}
	if string(data) != "# Note" {
		t.Errorf("content mismatch after move: %q", string(data))
	}
}

func TestCmdPropertySet(t *testing.T) {
	vaultDir := t.TempDir()

	content := "---\ntype: decision\nstatus: active\ncreated: 2024-01-15\n---\n\n# My Decision\n"
	notePath := filepath.Join(vaultDir, "My Decision.md")
	os.WriteFile(notePath, []byte(content), 0644)

	// Update existing property
	params := map[string]string{
		"file":  "My Decision",
		"name":  "status",
		"value": "archived",
	}
	if err := cmdPropertySet(vaultDir, params); err != nil {
		t.Fatalf("property:set: %v", err)
	}

	data, _ := os.ReadFile(notePath)
	if got := string(data); !contains(got, "status: archived") {
		t.Errorf("property not updated: %s", got)
	}
	if got := string(data); contains(got, "status: active") {
		t.Errorf("old property value still present: %s", got)
	}

	// Add new property
	params = map[string]string{
		"file":  "My Decision",
		"name":  "confidence",
		"value": "high",
	}
	if err := cmdPropertySet(vaultDir, params); err != nil {
		t.Fatalf("property:set (add): %v", err)
	}

	data, _ = os.ReadFile(notePath)
	if got := string(data); !contains(got, "confidence: high") {
		t.Errorf("new property not added: %s", got)
	}
}

func TestCmdSearch(t *testing.T) {
	vaultDir := t.TempDir()

	os.MkdirAll(filepath.Join(vaultDir, "decisions"), 0755)
	os.MkdirAll(filepath.Join(vaultDir, ".obsidian"), 0755)

	// Note with matching title
	os.WriteFile(filepath.Join(vaultDir, "decisions", "Paivot Architecture.md"),
		[]byte("# Architecture\nSome content."), 0644)

	// Note with matching content but not title
	os.WriteFile(filepath.Join(vaultDir, "decisions", "Other Decision.md"),
		[]byte("# Other\nThis relates to paivot infrastructure."), 0644)

	// Note that should not match
	os.WriteFile(filepath.Join(vaultDir, "decisions", "Unrelated.md"),
		[]byte("# Unrelated\nNothing here."), 0644)

	// Hidden note that should be skipped
	os.WriteFile(filepath.Join(vaultDir, ".obsidian", "paivot-config.md"),
		[]byte("# Config\npaivot settings."), 0644)

	params := map[string]string{"query": "paivot"}
	// cmdSearch writes to stdout; just verify no error
	if err := cmdSearch(vaultDir, params); err != nil {
		t.Fatalf("search: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
