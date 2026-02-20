package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- Unit Tests ---

func TestLoadBookmarks(t *testing.T) {
	vaultDir := t.TempDir()
	os.MkdirAll(filepath.Join(vaultDir, ".obsidian"), 0755)

	bm := bookmarksFile{
		Items: []bookmark{
			{Type: "file", Ctime: 1708300000000, Path: "folder/My Note.md"},
			{Type: "file", Ctime: 1708300000001, Path: "other/Note.md"},
		},
	}
	data, _ := json.Marshal(bm)
	os.WriteFile(filepath.Join(vaultDir, ".obsidian", "bookmarks.json"), data, 0644)

	loaded, err := loadBookmarks(vaultDir)
	if err != nil {
		t.Fatalf("loadBookmarks: %v", err)
	}
	if len(loaded.Items) != 2 {
		t.Fatalf("got %d items, want 2", len(loaded.Items))
	}
	if loaded.Items[0].Path != "folder/My Note.md" {
		t.Errorf("item[0].Path = %q, want %q", loaded.Items[0].Path, "folder/My Note.md")
	}
	if loaded.Items[1].Path != "other/Note.md" {
		t.Errorf("item[1].Path = %q, want %q", loaded.Items[1].Path, "other/Note.md")
	}
}

func TestLoadBookmarksEmpty(t *testing.T) {
	vaultDir := t.TempDir()

	// No .obsidian directory at all
	loaded, err := loadBookmarks(vaultDir)
	if err != nil {
		t.Fatalf("loadBookmarks on missing dir should not error: %v", err)
	}
	if len(loaded.Items) != 0 {
		t.Errorf("got %d items, want 0", len(loaded.Items))
	}
}

func TestLoadBookmarksNested(t *testing.T) {
	vaultDir := t.TempDir()
	os.MkdirAll(filepath.Join(vaultDir, ".obsidian"), 0755)

	bm := bookmarksFile{
		Items: []bookmark{
			{Type: "file", Ctime: 1708300000000, Path: "top-level.md"},
			{
				Type:  "group",
				Ctime: 1708300000001,
				Title: "My Group",
				Items: []bookmark{
					{Type: "file", Ctime: 1708300000002, Path: "nested/Note.md"},
					{Type: "file", Ctime: 1708300000003, Path: "nested/Other.md"},
				},
			},
		},
	}
	data, _ := json.Marshal(bm)
	os.WriteFile(filepath.Join(vaultDir, ".obsidian", "bookmarks.json"), data, 0644)

	loaded, err := loadBookmarks(vaultDir)
	if err != nil {
		t.Fatalf("loadBookmarks: %v", err)
	}
	if len(loaded.Items) != 2 {
		t.Fatalf("got %d top-level items, want 2", len(loaded.Items))
	}
	if loaded.Items[1].Type != "group" {
		t.Errorf("item[1].Type = %q, want group", loaded.Items[1].Type)
	}
	if len(loaded.Items[1].Items) != 2 {
		t.Errorf("group has %d items, want 2", len(loaded.Items[1].Items))
	}
}

func TestFlattenBookmarks(t *testing.T) {
	bm := bookmarksFile{
		Items: []bookmark{
			{Type: "file", Path: "top.md"},
			{
				Type:  "group",
				Title: "Group A",
				Items: []bookmark{
					{Type: "file", Path: "nested/a.md"},
					{
						Type:  "group",
						Title: "Sub Group",
						Items: []bookmark{
							{Type: "file", Path: "deep/b.md"},
						},
					},
				},
			},
			{Type: "folder", Path: "some-folder"},
			{Type: "search", Title: "my search"},
		},
	}

	paths := flattenBookmarks(bm.Items)
	if len(paths) != 3 {
		t.Fatalf("got %d paths, want 3 (only file-type)", len(paths))
	}
	want := []string{"top.md", "nested/a.md", "deep/b.md"}
	for i, w := range want {
		if paths[i] != w {
			t.Errorf("paths[%d] = %q, want %q", i, paths[i], w)
		}
	}
}

func TestAddBookmark(t *testing.T) {
	bm := &bookmarksFile{
		Items: []bookmark{
			{Type: "file", Path: "existing.md"},
		},
	}

	added := addBookmark(bm, "new/Note.md")
	if !added {
		t.Fatal("addBookmark should return true for new bookmark")
	}
	if len(bm.Items) != 2 {
		t.Fatalf("got %d items, want 2", len(bm.Items))
	}
	if bm.Items[1].Path != "new/Note.md" {
		t.Errorf("new item path = %q, want %q", bm.Items[1].Path, "new/Note.md")
	}
	if bm.Items[1].Type != "file" {
		t.Errorf("new item type = %q, want file", bm.Items[1].Type)
	}
	if bm.Items[1].Ctime == 0 {
		t.Error("new item ctime should be set")
	}
}

func TestAddBookmarkNoDuplicate(t *testing.T) {
	bm := &bookmarksFile{
		Items: []bookmark{
			{Type: "file", Path: "existing.md", Ctime: 1708300000000},
		},
	}

	added := addBookmark(bm, "existing.md")
	if added {
		t.Fatal("addBookmark should return false for duplicate")
	}
	if len(bm.Items) != 1 {
		t.Fatalf("got %d items, want 1 (no duplicate)", len(bm.Items))
	}
}

func TestRemoveBookmark(t *testing.T) {
	bm := &bookmarksFile{
		Items: []bookmark{
			{Type: "file", Path: "keep.md"},
			{Type: "file", Path: "remove.md"},
			{Type: "file", Path: "also-keep.md"},
		},
	}

	removed := removeBookmark(bm, "remove.md")
	if !removed {
		t.Fatal("removeBookmark should return true when bookmark found")
	}
	if len(bm.Items) != 2 {
		t.Fatalf("got %d items, want 2", len(bm.Items))
	}
	for _, item := range bm.Items {
		if item.Path == "remove.md" {
			t.Error("remove.md should not be present after removal")
		}
	}
}

func TestRemoveBookmarkFromGroup(t *testing.T) {
	bm := &bookmarksFile{
		Items: []bookmark{
			{Type: "file", Path: "top.md"},
			{
				Type:  "group",
				Title: "My Group",
				Items: []bookmark{
					{Type: "file", Path: "nested/keep.md"},
					{Type: "file", Path: "nested/remove.md"},
				},
			},
		},
	}

	removed := removeBookmark(bm, "nested/remove.md")
	if !removed {
		t.Fatal("removeBookmark should find bookmark in group")
	}
	if len(bm.Items[1].Items) != 1 {
		t.Fatalf("group has %d items, want 1", len(bm.Items[1].Items))
	}
	if bm.Items[1].Items[0].Path != "nested/keep.md" {
		t.Errorf("remaining group item = %q, want nested/keep.md", bm.Items[1].Items[0].Path)
	}
}

func TestRemoveBookmarkNotFound(t *testing.T) {
	bm := &bookmarksFile{
		Items: []bookmark{
			{Type: "file", Path: "keep.md"},
		},
	}

	removed := removeBookmark(bm, "nonexistent.md")
	if removed {
		t.Fatal("removeBookmark should return false when not found")
	}
}

// --- Integration Tests ---

func TestBookmarksListIntegration(t *testing.T) {
	vaultDir := t.TempDir()
	os.MkdirAll(filepath.Join(vaultDir, ".obsidian"), 0755)

	bm := bookmarksFile{
		Items: []bookmark{
			{Type: "file", Ctime: 1708300000000, Path: "notes/Alpha.md"},
			{Type: "file", Ctime: 1708300000001, Path: "notes/Beta.md"},
			{
				Type:  "group",
				Title: "Starred",
				Items: []bookmark{
					{Type: "file", Ctime: 1708300000002, Path: "notes/Gamma.md"},
				},
			},
		},
	}
	data, _ := json.Marshal(bm)
	os.WriteFile(filepath.Join(vaultDir, ".obsidian", "bookmarks.json"), data, 0644)

	got := captureStdout(func() {
		if err := cmdBookmarks(vaultDir, ""); err != nil {
			t.Fatalf("cmdBookmarks: %v", err)
		}
	})

	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3: %q", len(lines), got)
	}
	if lines[0] != "notes/Alpha.md" {
		t.Errorf("line[0] = %q, want notes/Alpha.md", lines[0])
	}
	if lines[1] != "notes/Beta.md" {
		t.Errorf("line[1] = %q, want notes/Beta.md", lines[1])
	}
	if lines[2] != "notes/Gamma.md" {
		t.Errorf("line[2] = %q, want notes/Gamma.md", lines[2])
	}
}

func TestBookmarksAddIntegration(t *testing.T) {
	vaultDir := t.TempDir()
	os.MkdirAll(filepath.Join(vaultDir, ".obsidian"), 0755)
	os.MkdirAll(filepath.Join(vaultDir, "notes"), 0755)

	// Create the note to bookmark
	os.WriteFile(filepath.Join(vaultDir, "notes", "MyNote.md"), []byte("# My Note\n"), 0644)

	// Create initial bookmarks file (empty)
	bm := bookmarksFile{Items: []bookmark{}}
	data, _ := json.Marshal(bm)
	os.WriteFile(filepath.Join(vaultDir, ".obsidian", "bookmarks.json"), data, 0644)

	// Add bookmark
	params := map[string]string{"file": "MyNote"}
	if err := cmdBookmarksAdd(vaultDir, params); err != nil {
		t.Fatalf("cmdBookmarksAdd: %v", err)
	}

	// Re-read the file and verify
	loaded, err := loadBookmarks(vaultDir)
	if err != nil {
		t.Fatalf("loadBookmarks: %v", err)
	}
	if len(loaded.Items) != 1 {
		t.Fatalf("got %d items, want 1", len(loaded.Items))
	}
	if loaded.Items[0].Path != "notes/MyNote.md" {
		t.Errorf("bookmark path = %q, want notes/MyNote.md", loaded.Items[0].Path)
	}
}

func TestBookmarksRemoveIntegration(t *testing.T) {
	vaultDir := t.TempDir()
	os.MkdirAll(filepath.Join(vaultDir, ".obsidian"), 0755)
	os.MkdirAll(filepath.Join(vaultDir, "notes"), 0755)

	// Create the note
	os.WriteFile(filepath.Join(vaultDir, "notes", "RemoveMe.md"), []byte("# Remove Me\n"), 0644)

	// Create bookmarks with the note
	bm := bookmarksFile{
		Items: []bookmark{
			{Type: "file", Ctime: 1708300000000, Path: "notes/RemoveMe.md"},
			{Type: "file", Ctime: 1708300000001, Path: "notes/KeepMe.md"},
		},
	}
	data, _ := json.Marshal(bm)
	os.WriteFile(filepath.Join(vaultDir, ".obsidian", "bookmarks.json"), data, 0644)

	// Remove bookmark
	params := map[string]string{"file": "RemoveMe"}
	if err := cmdBookmarksRemove(vaultDir, params); err != nil {
		t.Fatalf("cmdBookmarksRemove: %v", err)
	}

	// Verify it was removed
	loaded, err := loadBookmarks(vaultDir)
	if err != nil {
		t.Fatalf("loadBookmarks: %v", err)
	}
	if len(loaded.Items) != 1 {
		t.Fatalf("got %d items, want 1", len(loaded.Items))
	}
	if loaded.Items[0].Path != "notes/KeepMe.md" {
		t.Errorf("remaining item = %q, want notes/KeepMe.md", loaded.Items[0].Path)
	}
}

func TestBookmarksAddResolvesTitle(t *testing.T) {
	vaultDir := t.TempDir()
	os.MkdirAll(filepath.Join(vaultDir, ".obsidian"), 0755)
	os.MkdirAll(filepath.Join(vaultDir, "deep", "folder"), 0755)

	// Create a note in a nested folder
	os.WriteFile(
		filepath.Join(vaultDir, "deep", "folder", "Hidden Gem.md"),
		[]byte("---\naliases: [gem]\n---\n# Hidden Gem\n"),
		0644,
	)

	// Start with empty bookmarks
	bm := bookmarksFile{Items: []bookmark{}}
	data, _ := json.Marshal(bm)
	os.WriteFile(filepath.Join(vaultDir, ".obsidian", "bookmarks.json"), data, 0644)

	// Add by title
	params := map[string]string{"file": "Hidden Gem"}
	if err := cmdBookmarksAdd(vaultDir, params); err != nil {
		t.Fatalf("cmdBookmarksAdd: %v", err)
	}

	loaded, err := loadBookmarks(vaultDir)
	if err != nil {
		t.Fatalf("loadBookmarks: %v", err)
	}
	if len(loaded.Items) != 1 {
		t.Fatalf("got %d items, want 1", len(loaded.Items))
	}
	if loaded.Items[0].Path != "deep/folder/Hidden Gem.md" {
		t.Errorf("bookmark path = %q, want deep/folder/Hidden Gem.md", loaded.Items[0].Path)
	}
}

func TestBookmarksWithFormats(t *testing.T) {
	vaultDir := t.TempDir()
	os.MkdirAll(filepath.Join(vaultDir, ".obsidian"), 0755)

	bm := bookmarksFile{
		Items: []bookmark{
			{Type: "file", Ctime: 1708300000000, Path: "notes/Alpha.md"},
			{Type: "file", Ctime: 1708300000001, Path: "notes/Beta.md"},
		},
	}
	data, _ := json.Marshal(bm)
	os.WriteFile(filepath.Join(vaultDir, ".obsidian", "bookmarks.json"), data, 0644)

	// JSON
	jsonOut := captureStdout(func() {
		if err := cmdBookmarks(vaultDir, "json"); err != nil {
			t.Fatalf("json format: %v", err)
		}
	})
	var jsonParsed []string
	if err := json.Unmarshal([]byte(strings.TrimSpace(jsonOut)), &jsonParsed); err != nil {
		t.Fatalf("json parse error: %v\noutput: %q", err, jsonOut)
	}
	if len(jsonParsed) != 2 {
		t.Errorf("json: got %d items, want 2", len(jsonParsed))
	}

	// CSV
	csvOut := captureStdout(func() {
		if err := cmdBookmarks(vaultDir, "csv"); err != nil {
			t.Fatalf("csv format: %v", err)
		}
	})
	csvLines := strings.Split(strings.TrimSpace(csvOut), "\n")
	if len(csvLines) != 2 {
		t.Errorf("csv: got %d lines, want 2 (no header for list): %q", len(csvLines), csvOut)
	}

	// YAML
	yamlOut := captureStdout(func() {
		if err := cmdBookmarks(vaultDir, "yaml"); err != nil {
			t.Fatalf("yaml format: %v", err)
		}
	})
	if !strings.Contains(yamlOut, "- notes/Alpha.md") {
		t.Errorf("yaml: missing expected content, got: %q", yamlOut)
	}

	// TSV
	tsvOut := captureStdout(func() {
		if err := cmdBookmarks(vaultDir, "tsv"); err != nil {
			t.Fatalf("tsv format: %v", err)
		}
	})
	tsvLines := strings.Split(strings.TrimSpace(tsvOut), "\n")
	// TSV formatList has a header "file" then items
	if len(tsvLines) != 3 {
		t.Errorf("tsv: got %d lines, want 3 (header + 2 items): %q", len(tsvLines), tsvOut)
	}
	if tsvLines[0] != "file" {
		t.Errorf("tsv header = %q, want 'file'", tsvLines[0])
	}
}

func TestBookmarksAddCreatesObsidianDir(t *testing.T) {
	vaultDir := t.TempDir()
	// No .obsidian directory exists

	// Create the note
	os.WriteFile(filepath.Join(vaultDir, "NewNote.md"), []byte("# New Note\n"), 0644)

	// Add bookmark -- should create .obsidian/ and bookmarks.json
	params := map[string]string{"file": "NewNote"}
	if err := cmdBookmarksAdd(vaultDir, params); err != nil {
		t.Fatalf("cmdBookmarksAdd: %v", err)
	}

	// Verify .obsidian/bookmarks.json was created
	bmPath := filepath.Join(vaultDir, ".obsidian", "bookmarks.json")
	if _, err := os.Stat(bmPath); os.IsNotExist(err) {
		t.Fatal(".obsidian/bookmarks.json should have been created")
	}

	loaded, err := loadBookmarks(vaultDir)
	if err != nil {
		t.Fatalf("loadBookmarks: %v", err)
	}
	if len(loaded.Items) != 1 {
		t.Fatalf("got %d items, want 1", len(loaded.Items))
	}
	if loaded.Items[0].Path != "NewNote.md" {
		t.Errorf("bookmark path = %q, want NewNote.md", loaded.Items[0].Path)
	}
}

func TestBookmarksListNoObsidianDir(t *testing.T) {
	vaultDir := t.TempDir()
	// No .obsidian directory

	got := captureStdout(func() {
		if err := cmdBookmarks(vaultDir, ""); err != nil {
			t.Fatalf("cmdBookmarks should not error on missing dir: %v", err)
		}
	})

	// Should produce no output (empty list)
	if strings.TrimSpace(got) != "" {
		t.Errorf("expected empty output, got: %q", got)
	}
}

func TestBookmarksRemoveNoFile(t *testing.T) {
	vaultDir := t.TempDir()
	// No .obsidian directory

	// Create the note so resolveNote works
	os.WriteFile(filepath.Join(vaultDir, "Orphan.md"), []byte("# Orphan\n"), 0644)

	params := map[string]string{"file": "Orphan"}
	err := cmdBookmarksRemove(vaultDir, params)
	if err == nil {
		t.Fatal("cmdBookmarksRemove should error when bookmarks.json does not exist")
	}
	if !strings.Contains(err.Error(), "no bookmarks file") {
		t.Errorf("error = %q, want message about no bookmarks file", err.Error())
	}
}

func TestBookmarksAddDuplicateIntegration(t *testing.T) {
	vaultDir := t.TempDir()
	os.MkdirAll(filepath.Join(vaultDir, ".obsidian"), 0755)

	// Create the note
	os.WriteFile(filepath.Join(vaultDir, "Dup.md"), []byte("# Dup\n"), 0644)

	// Create bookmarks with this note already
	bm := bookmarksFile{
		Items: []bookmark{
			{Type: "file", Ctime: 1708300000000, Path: "Dup.md"},
		},
	}
	data, _ := json.Marshal(bm)
	os.WriteFile(filepath.Join(vaultDir, ".obsidian", "bookmarks.json"), data, 0644)

	// Add again -- should be a no-op
	params := map[string]string{"file": "Dup"}
	if err := cmdBookmarksAdd(vaultDir, params); err != nil {
		t.Fatalf("cmdBookmarksAdd duplicate: %v", err)
	}

	// Verify still only 1 item
	loaded, err := loadBookmarks(vaultDir)
	if err != nil {
		t.Fatalf("loadBookmarks: %v", err)
	}
	if len(loaded.Items) != 1 {
		t.Fatalf("got %d items, want 1 (no duplicate)", len(loaded.Items))
	}
}
