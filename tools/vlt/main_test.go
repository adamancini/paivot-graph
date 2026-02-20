package main

import (
	"os"
	"path/filepath"
	"strings"
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
			args:       []string{"vault=Claude", "search", "query=architecture"},
			wantCmd:    "search",
			wantParams: map[string]string{"vault": "Claude", "query": "architecture"},
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

func TestResolveNote_Alias(t *testing.T) {
	vaultDir := t.TempDir()

	os.MkdirAll(filepath.Join(vaultDir, "methodology"), 0755)
	os.WriteFile(
		filepath.Join(vaultDir, "methodology", "Sr PM Agent.md"),
		[]byte("---\naliases: [PM, Senior PM]\n---\n\n# Sr PM Agent\n"),
		0644,
	)

	// Resolve by alias
	path, err := resolveNote(vaultDir, "PM")
	if err != nil {
		t.Fatalf("alias resolution failed: %v", err)
	}
	relPath, _ := filepath.Rel(vaultDir, path)
	if relPath != "methodology/Sr PM Agent.md" {
		t.Errorf("got %q, want methodology/Sr PM Agent.md", relPath)
	}

	// Resolve by alias (case insensitive)
	path, err = resolveNote(vaultDir, "senior pm")
	if err != nil {
		t.Fatalf("case-insensitive alias failed: %v", err)
	}
	relPath, _ = filepath.Rel(vaultDir, path)
	if relPath != "methodology/Sr PM Agent.md" {
		t.Errorf("got %q, want methodology/Sr PM Agent.md", relPath)
	}

	// Filename match still takes priority
	path, err = resolveNote(vaultDir, "Sr PM Agent")
	if err != nil {
		t.Fatalf("filename resolution failed: %v", err)
	}
	relPath, _ = filepath.Rel(vaultDir, path)
	if relPath != "methodology/Sr PM Agent.md" {
		t.Errorf("got %q, want methodology/Sr PM Agent.md", relPath)
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

func TestCmdMove_RenameUpdatesLinks(t *testing.T) {
	vaultDir := t.TempDir()

	os.MkdirAll(filepath.Join(vaultDir, "_inbox"), 0755)
	os.MkdirAll(filepath.Join(vaultDir, "methodology"), 0755)

	// The note being renamed
	os.WriteFile(
		filepath.Join(vaultDir, "_inbox", "Old Name.md"),
		[]byte("# Old Name\n\nContent here.\n"),
		0644,
	)

	// Another note that references it
	os.WriteFile(
		filepath.Join(vaultDir, "methodology", "Developer Agent.md"),
		[]byte("# Developer\n\nSee [[Old Name]] and [[Old Name#Section|details]].\n"),
		0644,
	)

	params := map[string]string{
		"path": "_inbox/Old Name.md",
		"to":   "decisions/New Name.md",
	}
	if err := cmdMove(vaultDir, params); err != nil {
		t.Fatalf("move: %v", err)
	}

	// Verify the referencing file was updated
	data, _ := os.ReadFile(filepath.Join(vaultDir, "methodology", "Developer Agent.md"))
	got := string(data)

	if contains(got, "[[Old Name]]") {
		t.Error("old wikilink [[Old Name]] still present")
	}
	if !contains(got, "[[New Name]]") {
		t.Error("new wikilink [[New Name]] not found")
	}
	if !contains(got, "[[New Name#Section|details]]") {
		t.Error("new wikilink [[New Name#Section|details]] not found")
	}
}

func TestCmdMove_FolderOnlyNoLinkUpdate(t *testing.T) {
	vaultDir := t.TempDir()

	os.MkdirAll(filepath.Join(vaultDir, "_inbox"), 0755)

	// The note being moved (same filename, different folder)
	os.WriteFile(
		filepath.Join(vaultDir, "_inbox", "Note.md"),
		[]byte("# Note\n"),
		0644,
	)

	// Another note referencing it
	os.WriteFile(
		filepath.Join(vaultDir, "Referrer.md"),
		[]byte("See [[Note]] here.\n"),
		0644,
	)

	params := map[string]string{
		"path": "_inbox/Note.md",
		"to":   "decisions/Note.md",
	}
	if err := cmdMove(vaultDir, params); err != nil {
		t.Fatalf("move: %v", err)
	}

	// Link should remain unchanged (title didn't change)
	data, _ := os.ReadFile(filepath.Join(vaultDir, "Referrer.md"))
	if string(data) != "See [[Note]] here.\n" {
		t.Errorf("referrer was unexpectedly modified: %q", string(data))
	}
}

func TestCmdMove_UpdatesMdLinks(t *testing.T) {
	vaultDir := t.TempDir()

	os.MkdirAll(filepath.Join(vaultDir, "_inbox"), 0755)

	// The note being moved
	os.WriteFile(
		filepath.Join(vaultDir, "_inbox", "Note.md"),
		[]byte("# Note\n"),
		0644,
	)

	// Another note referencing it via markdown link
	os.WriteFile(
		filepath.Join(vaultDir, "Referrer.md"),
		[]byte("See [note](_inbox/Note.md) and [heading](_inbox/Note.md#section) here.\n"),
		0644,
	)

	params := map[string]string{
		"path": "_inbox/Note.md",
		"to":   "decisions/Note.md",
	}
	if err := cmdMove(vaultDir, params); err != nil {
		t.Fatalf("move: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(vaultDir, "Referrer.md"))
	got := string(data)

	if strings.Contains(got, "_inbox/Note.md") {
		t.Error("old markdown link path still present")
	}
	if !strings.Contains(got, "decisions/Note.md") {
		t.Error("new markdown link path not found")
	}
	if !strings.Contains(got, "decisions/Note.md#section") {
		t.Error("markdown link fragment not preserved")
	}
}

func TestCmdBacklinks(t *testing.T) {
	vaultDir := t.TempDir()

	os.MkdirAll(filepath.Join(vaultDir, "methodology"), 0755)

	os.WriteFile(
		filepath.Join(vaultDir, "methodology", "Developer Agent.md"),
		[]byte("Read [[Session Operating Mode]] first.\n"),
		0644,
	)
	os.WriteFile(
		filepath.Join(vaultDir, "methodology", "Retro Agent.md"),
		[]byte("# Retro\n\nNo links to SOM.\n"),
		0644,
	)

	// Just verify no error (output goes to stdout)
	params := map[string]string{"file": "Session Operating Mode"}
	if err := cmdBacklinks(vaultDir, params, ""); err != nil {
		t.Fatalf("backlinks: %v", err)
	}
}

func TestCmdLinks(t *testing.T) {
	vaultDir := t.TempDir()

	os.MkdirAll(filepath.Join(vaultDir, "methodology"), 0755)

	// Target note with outgoing links
	os.WriteFile(
		filepath.Join(vaultDir, "methodology", "Developer Agent.md"),
		[]byte("# Developer\n\nSee [[Session Operating Mode]] and [[Nonexistent Note]].\n"),
		0644,
	)

	// One of the linked notes exists
	os.WriteFile(
		filepath.Join(vaultDir, "Session Operating Mode.md"),
		[]byte("# SOM\n"),
		0644,
	)

	// Just verify no error (output goes to stdout)
	params := map[string]string{"file": "Developer Agent"}
	if err := cmdLinks(vaultDir, params, ""); err != nil {
		t.Fatalf("links: %v", err)
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
	os.WriteFile(filepath.Join(vaultDir, "decisions", "System Architecture.md"),
		[]byte("# Architecture\nSome content."), 0644)

	// Note with matching content but not title
	os.WriteFile(filepath.Join(vaultDir, "decisions", "Other Decision.md"),
		[]byte("# Other\nThis relates to system infrastructure."), 0644)

	// Note that should not match
	os.WriteFile(filepath.Join(vaultDir, "decisions", "Unrelated.md"),
		[]byte("# Unrelated\nNothing here."), 0644)

	// Hidden note that should be skipped
	os.WriteFile(filepath.Join(vaultDir, ".obsidian", "system-config.md"),
		[]byte("# Config\nsystem settings."), 0644)

	params := map[string]string{"query": "system"}
	// cmdSearch writes to stdout; just verify no error
	if err := cmdSearch(vaultDir, params, ""); err != nil {
		t.Fatalf("search: %v", err)
	}
}

func TestParseSearchQuery(t *testing.T) {
	tests := []struct {
		query       string
		wantText    string
		wantFilters map[string]string
	}{
		{
			query:       "architecture",
			wantText:    "architecture",
			wantFilters: map[string]string{},
		},
		{
			query:       "architecture [status:active]",
			wantText:    "architecture",
			wantFilters: map[string]string{"status": "active"},
		},
		{
			query:       "[status:active] [type:decision]",
			wantText:    "",
			wantFilters: map[string]string{"status": "active", "type": "decision"},
		},
		{
			query:       "search term [status:active] more text",
			wantText:    "search term  more text",
			wantFilters: map[string]string{"status": "active"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			text, filters := parseSearchQuery(tt.query)
			if text != tt.wantText {
				t.Errorf("text = %q, want %q", text, tt.wantText)
			}
			if len(filters) != len(tt.wantFilters) {
				t.Errorf("got %d filters, want %d", len(filters), len(tt.wantFilters))
			}
			for k, v := range tt.wantFilters {
				if filters[k] != v {
					t.Errorf("filter[%q] = %q, want %q", k, filters[k], v)
				}
			}
		})
	}
}

func TestCmdSearch_PropertyFilter(t *testing.T) {
	vaultDir := t.TempDir()

	os.MkdirAll(filepath.Join(vaultDir, "decisions"), 0755)

	os.WriteFile(filepath.Join(vaultDir, "decisions", "Active Decision.md"),
		[]byte("---\ntype: decision\nstatus: active\n---\n\n# Active\nSome content."), 0644)

	os.WriteFile(filepath.Join(vaultDir, "decisions", "Archived Decision.md"),
		[]byte("---\ntype: decision\nstatus: archived\n---\n\n# Archived\nOther content."), 0644)

	os.WriteFile(filepath.Join(vaultDir, "decisions", "No Frontmatter.md"),
		[]byte("# No FM\nPlain note."), 0644)

	// Filter by status:active should find only the active note
	params := map[string]string{"query": "[status:active]"}
	// Just verify no error; output goes to stdout
	if err := cmdSearch(vaultDir, params, ""); err != nil {
		t.Fatalf("search with property filter: %v", err)
	}
}

func TestCmdSearch_PropertyFilterWithText(t *testing.T) {
	vaultDir := t.TempDir()

	os.WriteFile(filepath.Join(vaultDir, "Match.md"),
		[]byte("---\nstatus: active\n---\n\n# Match\narchitecture discussion."), 0644)

	os.WriteFile(filepath.Join(vaultDir, "NoMatch.md"),
		[]byte("---\nstatus: archived\n---\n\n# NoMatch\narchitecture discussion."), 0644)

	params := map[string]string{"query": "architecture [status:active]"}
	if err := cmdSearch(vaultDir, params, ""); err != nil {
		t.Fatalf("search with text + filter: %v", err)
	}
}

func TestCmdSearch_MultipleFilters(t *testing.T) {
	vaultDir := t.TempDir()

	os.WriteFile(filepath.Join(vaultDir, "Both.md"),
		[]byte("---\ntype: decision\nstatus: active\n---\n\n# Both\nContent."), 0644)

	os.WriteFile(filepath.Join(vaultDir, "OneOnly.md"),
		[]byte("---\ntype: pattern\nstatus: active\n---\n\n# OneOnly\nContent."), 0644)

	params := map[string]string{"query": "[type:decision] [status:active]"}
	if err := cmdSearch(vaultDir, params, ""); err != nil {
		t.Fatalf("search with multiple filters: %v", err)
	}
}

func TestCmdPrepend(t *testing.T) {
	vaultDir := t.TempDir()

	// With frontmatter: should insert after ---
	os.WriteFile(
		filepath.Join(vaultDir, "WithFM.md"),
		[]byte("---\ntype: note\n---\n\n# Existing Content\n"),
		0644,
	)

	params := map[string]string{"file": "WithFM", "content": "PREPENDED\n"}
	if err := cmdPrepend(vaultDir, params); err != nil {
		t.Fatalf("prepend with FM: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(vaultDir, "WithFM.md"))
	got := string(data)
	want := "---\ntype: note\n---\nPREPENDED\n\n# Existing Content\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	// Without frontmatter: should insert at top
	os.WriteFile(
		filepath.Join(vaultDir, "NoFM.md"),
		[]byte("# Existing Content\n"),
		0644,
	)

	params = map[string]string{"file": "NoFM", "content": "TOP\n"}
	if err := cmdPrepend(vaultDir, params); err != nil {
		t.Fatalf("prepend without FM: %v", err)
	}

	data, _ = os.ReadFile(filepath.Join(vaultDir, "NoFM.md"))
	got = string(data)
	want = "TOP\n# Existing Content\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCmdDelete_Trash(t *testing.T) {
	vaultDir := t.TempDir()

	notePath := filepath.Join(vaultDir, "ToTrash.md")
	os.WriteFile(notePath, []byte("# Delete me\n"), 0644)

	params := map[string]string{"file": "ToTrash"}
	if err := cmdDelete(vaultDir, params, false); err != nil {
		t.Fatalf("delete (trash): %v", err)
	}

	// Original should be gone
	if _, err := os.Stat(notePath); !os.IsNotExist(err) {
		t.Error("original file still exists after trash")
	}

	// Should exist in .trash
	trashPath := filepath.Join(vaultDir, ".trash", "ToTrash.md")
	if _, err := os.Stat(trashPath); os.IsNotExist(err) {
		t.Error("file not found in .trash")
	}
}

func TestCmdDelete_Permanent(t *testing.T) {
	vaultDir := t.TempDir()

	notePath := filepath.Join(vaultDir, "ToDelete.md")
	os.WriteFile(notePath, []byte("# Delete me\n"), 0644)

	params := map[string]string{"file": "ToDelete"}
	if err := cmdDelete(vaultDir, params, true); err != nil {
		t.Fatalf("delete (permanent): %v", err)
	}

	if _, err := os.Stat(notePath); !os.IsNotExist(err) {
		t.Error("file still exists after permanent delete")
	}

	// Should NOT exist in .trash
	trashPath := filepath.Join(vaultDir, ".trash", "ToDelete.md")
	if _, err := os.Stat(trashPath); !os.IsNotExist(err) {
		t.Error("file unexpectedly found in .trash after permanent delete")
	}
}

func TestCmdProperties(t *testing.T) {
	vaultDir := t.TempDir()

	os.WriteFile(
		filepath.Join(vaultDir, "Props.md"),
		[]byte("---\ntype: decision\nstatus: active\n---\n\n# Note\n"),
		0644,
	)

	// Just verify no error (output goes to stdout)
	params := map[string]string{"file": "Props"}
	if err := cmdProperties(vaultDir, params, ""); err != nil {
		t.Fatalf("properties: %v", err)
	}
}

func TestCmdPropertyRemove(t *testing.T) {
	vaultDir := t.TempDir()

	notePath := filepath.Join(vaultDir, "Note.md")
	os.WriteFile(notePath, []byte("---\ntype: decision\nstatus: active\ncreated: 2024-01-15\n---\n\n# Note\n"), 0644)

	params := map[string]string{"file": "Note", "name": "status"}
	if err := cmdPropertyRemove(vaultDir, params); err != nil {
		t.Fatalf("property:remove: %v", err)
	}

	data, _ := os.ReadFile(notePath)
	got := string(data)

	if contains(got, "status:") {
		t.Error("property 'status' still present after removal")
	}
	if !contains(got, "type: decision") || !contains(got, "created: 2024-01-15") {
		t.Error("other properties were affected by removal")
	}
}

func TestCmdOrphans(t *testing.T) {
	vaultDir := t.TempDir()

	// A references B; C is orphaned
	os.WriteFile(
		filepath.Join(vaultDir, "A.md"),
		[]byte("# A\n\nSee [[B]] for details.\n"),
		0644,
	)
	os.WriteFile(
		filepath.Join(vaultDir, "B.md"),
		[]byte("# B\n\nReferenced by A.\n"),
		0644,
	)
	os.WriteFile(
		filepath.Join(vaultDir, "C.md"),
		[]byte("# C\n\nNobody links to me.\n"),
		0644,
	)

	// Just verify no error
	if err := cmdOrphans(vaultDir, ""); err != nil {
		t.Fatalf("orphans: %v", err)
	}
}

func TestCmdOrphans_AliasAware(t *testing.T) {
	vaultDir := t.TempDir()

	// A references "Alt Name" which is an alias of B
	os.WriteFile(
		filepath.Join(vaultDir, "A.md"),
		[]byte("# A\n\nSee [[Alt Name]].\n"),
		0644,
	)
	os.WriteFile(
		filepath.Join(vaultDir, "B.md"),
		[]byte("---\naliases: [Alt Name]\n---\n\n# B\n"),
		0644,
	)
	os.WriteFile(
		filepath.Join(vaultDir, "C.md"),
		[]byte("# C\n\nOrphan.\n"),
		0644,
	)

	// Just verify no error (A is orphaned since nothing links to it,
	// B is NOT orphaned due to alias, C is orphaned)
	if err := cmdOrphans(vaultDir, ""); err != nil {
		t.Fatalf("orphans: %v", err)
	}
}

func TestCmdUnresolved(t *testing.T) {
	vaultDir := t.TempDir()

	os.WriteFile(
		filepath.Join(vaultDir, "Existing.md"),
		[]byte("# Existing\n"),
		0644,
	)
	os.WriteFile(
		filepath.Join(vaultDir, "Referrer.md"),
		[]byte("# Referrer\n\n[[Existing]] and [[Ghost Note]] and ![[Missing Embed]].\n"),
		0644,
	)

	// Just verify no error
	if err := cmdUnresolved(vaultDir, ""); err != nil {
		t.Fatalf("unresolved: %v", err)
	}
}

func TestCmdFiles(t *testing.T) {
	vaultDir := t.TempDir()

	os.MkdirAll(filepath.Join(vaultDir, "sub"), 0755)
	os.MkdirAll(filepath.Join(vaultDir, ".obsidian"), 0755)
	os.WriteFile(filepath.Join(vaultDir, "root.md"), []byte("# Root\n"), 0644)
	os.WriteFile(filepath.Join(vaultDir, "sub", "child.md"), []byte("# Child\n"), 0644)
	os.WriteFile(filepath.Join(vaultDir, ".obsidian", "config.md"), []byte("hidden\n"), 0644)

	// List all
	params := map[string]string{}
	if err := cmdFiles(vaultDir, params, false, ""); err != nil {
		t.Fatalf("files: %v", err)
	}

	// Total count
	if err := cmdFiles(vaultDir, params, true, ""); err != nil {
		t.Fatalf("files total: %v", err)
	}

	// Filter by folder
	params = map[string]string{"folder": "sub"}
	if err := cmdFiles(vaultDir, params, false, ""); err != nil {
		t.Fatalf("files folder: %v", err)
	}
}

// ---------------------------------------------------------------------------
// write command tests
// ---------------------------------------------------------------------------

// Unit test 1: write replaces body while preserving frontmatter
func TestCmdWriteReplacesBody(t *testing.T) {
	vaultDir := t.TempDir()

	original := "---\ntype: decision\nstatus: active\n---\n\n# Old Body\n\nOld content here.\n"
	notePath := filepath.Join(vaultDir, "Note.md")
	os.WriteFile(notePath, []byte(original), 0644)

	params := map[string]string{
		"file":    "Note",
		"content": "# New Body\n\nCompletely replaced.\n",
	}
	if err := cmdWrite(vaultDir, params); err != nil {
		t.Fatalf("write: %v", err)
	}

	data, _ := os.ReadFile(notePath)
	got := string(data)

	// Frontmatter must be preserved
	if !strings.Contains(got, "type: decision") {
		t.Error("frontmatter property 'type' lost after write")
	}
	if !strings.Contains(got, "status: active") {
		t.Error("frontmatter property 'status' lost after write")
	}

	// Body must be replaced
	if strings.Contains(got, "Old Body") {
		t.Error("old body content still present after write")
	}
	if !strings.Contains(got, "Completely replaced.") {
		t.Error("new body content not found after write")
	}
}

// Unit test 2: write to note without frontmatter replaces entire content
func TestCmdWriteNoFrontmatter(t *testing.T) {
	vaultDir := t.TempDir()

	original := "# Old Title\n\nSome old content.\n"
	notePath := filepath.Join(vaultDir, "Plain.md")
	os.WriteFile(notePath, []byte(original), 0644)

	params := map[string]string{
		"file":    "Plain",
		"content": "# New Title\n\nNew content.\n",
	}
	if err := cmdWrite(vaultDir, params); err != nil {
		t.Fatalf("write: %v", err)
	}

	data, _ := os.ReadFile(notePath)
	got := string(data)

	if strings.Contains(got, "Old Title") {
		t.Error("old content still present in note without frontmatter")
	}
	if got != "# New Title\n\nNew content.\n" {
		t.Errorf("unexpected content: %q", got)
	}
}

// Unit test 3: write empty content results in frontmatter-only note
func TestCmdWriteEmptyBody(t *testing.T) {
	vaultDir := t.TempDir()

	original := "---\ntype: note\n---\n\n# Content\n"
	notePath := filepath.Join(vaultDir, "EmptyBody.md")
	os.WriteFile(notePath, []byte(original), 0644)

	params := map[string]string{
		"file":    "EmptyBody",
		"content": "",
	}
	if err := cmdWrite(vaultDir, params); err != nil {
		t.Fatalf("write: %v", err)
	}

	data, _ := os.ReadFile(notePath)
	got := string(data)

	// Should have frontmatter but no body content
	if !strings.Contains(got, "---\ntype: note\n---") {
		t.Error("frontmatter lost when writing empty body")
	}
	if strings.Contains(got, "# Content") {
		t.Error("old body still present after writing empty content")
	}
}

// Unit test 4: write without file= returns error
func TestCmdWriteRequiresFile(t *testing.T) {
	vaultDir := t.TempDir()

	params := map[string]string{
		"content": "some content",
	}
	err := cmdWrite(vaultDir, params)
	if err == nil {
		t.Fatal("expected error when file= not provided")
	}
	if !strings.Contains(err.Error(), "file") {
		t.Errorf("error should mention 'file', got: %v", err)
	}
}

// Unit test 5: write to nonexistent note returns error
func TestCmdWriteNoteNotFound(t *testing.T) {
	vaultDir := t.TempDir()

	params := map[string]string{
		"file":    "Nonexistent",
		"content": "some content",
	}
	err := cmdWrite(vaultDir, params)
	if err == nil {
		t.Fatal("expected error for nonexistent note")
	}
}

// ---------------------------------------------------------------------------
// Integration tests (real files, no mocks)
// ---------------------------------------------------------------------------

// Integration test 6: create real note with frontmatter + body, write new body, verify frontmatter intact
func TestWritePreservesFrontmatter(t *testing.T) {
	vaultDir := t.TempDir()
	os.MkdirAll(filepath.Join(vaultDir, "decisions"), 0755)

	original := "---\ntype: decision\nstatus: active\ncreated: 2026-02-19\naliases: [Dec1, First Decision]\n---\n\n# Original Decision\n\nOriginal body with [[wikilinks]] and content.\n"
	notePath := filepath.Join(vaultDir, "decisions", "My Decision.md")
	os.WriteFile(notePath, []byte(original), 0644)

	params := map[string]string{
		"file":    "My Decision",
		"content": "# Updated Decision\n\nNew body with different content.\n",
	}
	if err := cmdWrite(vaultDir, params); err != nil {
		t.Fatalf("write: %v", err)
	}

	data, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("failed to read back note: %v", err)
	}
	got := string(data)

	// All frontmatter properties must be intact
	if !strings.Contains(got, "type: decision") {
		t.Error("frontmatter 'type' lost")
	}
	if !strings.Contains(got, "status: active") {
		t.Error("frontmatter 'status' lost")
	}
	if !strings.Contains(got, "created: 2026-02-19") {
		t.Error("frontmatter 'created' lost")
	}
	if !strings.Contains(got, "aliases: [Dec1, First Decision]") {
		t.Error("frontmatter 'aliases' lost")
	}

	// New body must be present
	if !strings.Contains(got, "# Updated Decision") {
		t.Error("new body not found")
	}
	if !strings.Contains(got, "New body with different content.") {
		t.Error("new body content not found")
	}

	// Old body must be gone
	if strings.Contains(got, "Original Decision") {
		t.Error("old body content still present")
	}
	if strings.Contains(got, "[[wikilinks]]") {
		t.Error("old wikilinks still present in body")
	}
}

// Integration test 7: write content piped from stdin (test the stdin fallback path)
// Note: We cannot truly pipe stdin in a test, but we can test the code path
// by passing content="" and verifying behavior. The actual stdin path is tested
// by verifying the function signature accepts empty content gracefully when
// there's no piped input. Instead, we test that content= takes priority.
func TestWriteViaContentParam(t *testing.T) {
	vaultDir := t.TempDir()

	notePath := filepath.Join(vaultDir, "StdinNote.md")
	os.WriteFile(notePath, []byte("---\ntitle: stdin test\n---\n\nOld body.\n"), 0644)

	params := map[string]string{
		"file":    "StdinNote",
		"content": "Body from content param.\n",
	}
	if err := cmdWrite(vaultDir, params); err != nil {
		t.Fatalf("write: %v", err)
	}

	data, _ := os.ReadFile(notePath)
	got := string(data)

	if !strings.Contains(got, "Body from content param.") {
		t.Error("content= param not applied")
	}
	if strings.Contains(got, "Old body.") {
		t.Error("old body still present")
	}
}

// Integration test 8: write content then read back with cmdRead to verify round-trip
func TestWriteThenRead(t *testing.T) {
	vaultDir := t.TempDir()

	notePath := filepath.Join(vaultDir, "RoundTrip.md")
	os.WriteFile(notePath, []byte("---\ntype: test\n---\n\n# Before\n"), 0644)

	newBody := "# After Write\n\nThis is the new content.\n"
	writeParams := map[string]string{
		"file":    "RoundTrip",
		"content": newBody,
	}
	if err := cmdWrite(vaultDir, writeParams); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Read back with resolveNote (same path cmdRead uses)
	path, err := resolveNote(vaultDir, "RoundTrip")
	if err != nil {
		t.Fatalf("resolveNote: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("readFile: %v", err)
	}

	got := string(data)
	if !strings.Contains(got, "type: test") {
		t.Error("frontmatter not preserved on read-back")
	}
	if !strings.Contains(got, "# After Write") {
		t.Error("new body not found on read-back")
	}
	if !strings.Contains(got, "This is the new content.") {
		t.Error("new body content not found on read-back")
	}
}

// Integration test 9: write to nonexistent file returns error, file does not appear
func TestWriteDoesNotCreateFile(t *testing.T) {
	vaultDir := t.TempDir()

	params := map[string]string{
		"file":    "Ghost Note",
		"content": "Should not be created",
	}
	err := cmdWrite(vaultDir, params)
	if err == nil {
		t.Fatal("expected error for nonexistent note")
	}

	// Verify no file was created
	matches, _ := filepath.Glob(filepath.Join(vaultDir, "*.md"))
	if len(matches) > 0 {
		t.Errorf("unexpected files created: %v", matches)
	}
}

// E2E test 10: full workflow -- create vault, create note with frontmatter and body,
// run cmdWrite with new content, verify with cmdRead and cmdProperties
func TestE2EWriteCommand(t *testing.T) {
	vaultDir := t.TempDir()
	os.MkdirAll(filepath.Join(vaultDir, "methodology"), 0755)

	// Step 1: Create a note with frontmatter and body
	originalContent := "---\ntype: methodology\nstatus: active\ncreated: 2026-02-19\n---\n\n# Original Heading\n\nOriginal body paragraph.\n\n## Section 2\n\nMore original content.\n"
	notePath := filepath.Join(vaultDir, "methodology", "Test Method.md")
	os.WriteFile(notePath, []byte(originalContent), 0644)

	// Step 2: Write new body content
	newBody := "# Revised Heading\n\nCompletely new body.\n\n## New Section\n\nAll new content here.\n"
	writeParams := map[string]string{
		"file":    "Test Method",
		"content": newBody,
	}
	if err := cmdWrite(vaultDir, writeParams); err != nil {
		t.Fatalf("E2E write: %v", err)
	}

	// Step 3: Verify with direct file read (simulates cmdRead)
	data, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("E2E read: %v", err)
	}
	got := string(data)

	// Frontmatter must be fully preserved
	if !strings.Contains(got, "type: methodology") {
		t.Error("E2E: frontmatter 'type' missing")
	}
	if !strings.Contains(got, "status: active") {
		t.Error("E2E: frontmatter 'status' missing")
	}
	if !strings.Contains(got, "created: 2026-02-19") {
		t.Error("E2E: frontmatter 'created' missing")
	}

	// New body must be present
	if !strings.Contains(got, "# Revised Heading") {
		t.Error("E2E: new heading not found")
	}
	if !strings.Contains(got, "All new content here.") {
		t.Error("E2E: new body content not found")
	}

	// Old body must be gone
	if strings.Contains(got, "Original Heading") {
		t.Error("E2E: old heading still present")
	}
	if strings.Contains(got, "Original body paragraph") {
		t.Error("E2E: old body still present")
	}

	// Step 4: Verify properties are intact via extractFrontmatter
	yaml, _, hasFM := extractFrontmatter(got)
	if !hasFM {
		t.Fatal("E2E: no frontmatter found after write")
	}
	typeVal, ok := frontmatterGetValue(yaml, "type")
	if !ok || typeVal != "methodology" {
		t.Errorf("E2E: type property = %q, want 'methodology'", typeVal)
	}
	statusVal, ok := frontmatterGetValue(yaml, "status")
	if !ok || statusVal != "active" {
		t.Errorf("E2E: status property = %q, want 'active'", statusVal)
	}
	createdVal, ok := frontmatterGetValue(yaml, "created")
	if !ok || createdVal != "2026-02-19" {
		t.Errorf("E2E: created property = %q, want '2026-02-19'", createdVal)
	}

	// Step 5: Verify the complete structure (frontmatter + separator + body)
	expectedPrefix := "---\ntype: methodology\nstatus: active\ncreated: 2026-02-19\n---\n"
	if !strings.HasPrefix(got, expectedPrefix) {
		t.Errorf("E2E: file does not start with expected frontmatter block.\nGot prefix: %q", got[:min(len(got), len(expectedPrefix)+20)])
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
