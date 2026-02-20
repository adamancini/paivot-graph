package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// --- Unit Tests ---

func TestMaskFencedCodeBlock(t *testing.T) {
	input := "Before\n```\n[[Link]] and #tag\n```\nAfter"
	got := maskInertContent(input)

	if strings.Contains(got, "[[Link]]") {
		t.Error("wikilink inside fenced code block should be masked")
	}
	if strings.Contains(got, "#tag") {
		t.Error("tag inside fenced code block should be masked")
	}
	if !strings.HasPrefix(got, "Before\n") {
		t.Error("content before fence should be unchanged")
	}
	if !strings.HasSuffix(got, "\nAfter") {
		t.Error("content after fence should be unchanged")
	}
}

func TestMaskFencedCodeBlockWithLanguage(t *testing.T) {
	languages := []string{"go", "python", "javascript", "rust", "yaml"}
	for _, lang := range languages {
		t.Run(lang, func(t *testing.T) {
			input := "Text\n```" + lang + "\n[[Link]] #tag\n```\nMore text"
			got := maskInertContent(input)

			if strings.Contains(got, "[[Link]]") {
				t.Errorf("wikilink inside ```%s block should be masked", lang)
			}
			if strings.Contains(got, "#tag") {
				t.Errorf("tag inside ```%s block should be masked", lang)
			}
			// Fence delimiter itself is NOT masked
			if !strings.Contains(got, "```"+lang) {
				t.Errorf("fence delimiter ```%s should be preserved", lang)
			}
		})
	}
}

func TestMaskMermaidBlock(t *testing.T) {
	input := "Before\n```mermaid\ngraph TD\nA[[Node A]] --> B[[Node B]]\n```\nAfter"
	got := maskInertContent(input)

	if strings.Contains(got, "[[Node A]]") {
		t.Error("wikilink inside mermaid block should be masked")
	}
	if strings.Contains(got, "[[Node B]]") {
		t.Error("second wikilink inside mermaid block should be masked")
	}
	if !strings.Contains(got, "```mermaid") {
		t.Error("mermaid fence delimiter should be preserved")
	}
}

func TestMaskPreservesLength(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "basic fenced block",
			input: "Before\n```\n[[Link]] and #tag\n```\nAfter",
		},
		{
			name:  "language tagged block",
			input: "Text\n```go\nfunc main() { fmt.Println(\"[[Link]]\") }\n```\nEnd",
		},
		{
			name:  "mermaid block",
			input: "```mermaid\nA[[X]] --> B[[Y]]\n```",
		},
		{
			name:  "multiple blocks",
			input: "```\nblock1\n```\nMiddle\n```python\nblock2\n```",
		},
		{
			name:  "empty block",
			input: "```\n```",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := maskInertContent(tt.input)
			if len(got) != len(tt.input) {
				t.Errorf("length changed: input=%d, output=%d", len(tt.input), len(got))
			}
		})
	}
}

func TestMaskPreservesNewlines(t *testing.T) {
	input := "Before\n```\nline1\nline2\nline3\n```\nAfter"
	got := maskInertContent(input)

	inputNewlines := strings.Count(input, "\n")
	gotNewlines := strings.Count(got, "\n")

	if inputNewlines != gotNewlines {
		t.Errorf("newline count changed: input=%d, output=%d", inputNewlines, gotNewlines)
	}

	// Verify specific newlines within masked zone are preserved
	lines := strings.Split(got, "\n")
	// lines: "Before", "```", "     ", "     ", "     ", "```", "After"
	if len(lines) != 7 {
		t.Fatalf("expected 7 lines, got %d", len(lines))
	}
	// Masked lines should be all spaces (non-newline chars replaced)
	for i := 2; i <= 4; i++ {
		if strings.TrimRight(lines[i], " ") != "" {
			t.Errorf("line %d should be all spaces, got %q", i, lines[i])
		}
	}
}

func TestMaskNonFencedContentUnchanged(t *testing.T) {
	input := "# Title\n\nSome [[Link]] and #tag text.\n\nMore content."
	got := maskInertContent(input)

	if got != input {
		t.Errorf("non-fenced content should be unchanged:\ngot:  %q\nwant: %q", got, input)
	}
}

func TestMaskMultipleFencedBlocks(t *testing.T) {
	input := "Start\n```\n[[A]]\n```\nMiddle [[B]]\n```go\n[[C]] #tag\n```\nEnd"
	got := maskInertContent(input)

	if strings.Contains(got, "[[A]]") {
		t.Error("wikilink in first fenced block should be masked")
	}
	if !strings.Contains(got, "[[B]]") {
		t.Error("wikilink between fenced blocks should be preserved")
	}
	if strings.Contains(got, "[[C]]") {
		t.Error("wikilink in second fenced block should be masked")
	}
}

func TestMaskUnclosedFence(t *testing.T) {
	input := "Before\n```\n[[Link]] and #tag\nmore content"
	got := maskInertContent(input)

	if strings.Contains(got, "[[Link]]") {
		t.Error("wikilink after unclosed fence should be masked (Obsidian behavior)")
	}
	if strings.Contains(got, "#tag") {
		t.Error("tag after unclosed fence should be masked")
	}
	if len(got) != len(input) {
		t.Errorf("length changed: input=%d, output=%d", len(input), len(got))
	}
}

func TestMaskNestedBackticks(t *testing.T) {
	input := "```\nSome `inline` code with [[Link]]\n```"
	got := maskInertContent(input)

	if strings.Contains(got, "[[Link]]") {
		t.Error("wikilink inside fenced block with inline backticks should be masked")
	}
}

func TestMaskEmptyFencedBlock(t *testing.T) {
	input := "Before\n```\n```\nAfter"
	got := maskInertContent(input)

	if len(got) != len(input) {
		t.Errorf("length changed: input=%d, output=%d", len(input), len(got))
	}
	if !strings.Contains(got, "Before") || !strings.Contains(got, "After") {
		t.Error("content outside empty fenced block should be unchanged")
	}
}

func TestRegisteredPassesPattern(t *testing.T) {
	// Save and restore global state
	origPasses := make([]maskPass, len(inertPasses))
	copy(origPasses, inertPasses)
	defer func() { inertPasses = origPasses }()

	// Clear passes and verify
	inertPasses = nil

	var callOrder []int

	registerMaskPass(func(text string) string {
		callOrder = append(callOrder, 1)
		return strings.ReplaceAll(text, "AAA", "BBB")
	})
	registerMaskPass(func(text string) string {
		callOrder = append(callOrder, 2)
		return strings.ReplaceAll(text, "BBB", "CCC")
	})

	result := maskInertContent("AAA")

	if result != "CCC" {
		t.Errorf("passes not applied in order: got %q, want %q", result, "CCC")
	}
	if len(callOrder) != 2 || callOrder[0] != 1 || callOrder[1] != 2 {
		t.Errorf("pass execution order wrong: %v", callOrder)
	}
}

// --- Integration Tests ---

func TestParseWikilinksIgnoresFencedCode(t *testing.T) {
	text := "Normal [[Outside]] link.\n```\n[[Inside]] should be ignored.\n```\nMore [[AlsoOutside]]."
	masked := maskInertContent(text)
	links := parseWikilinks(masked)

	titles := make(map[string]bool)
	for _, l := range links {
		titles[l.Title] = true
	}

	if !titles["Outside"] {
		t.Error("expected to find [[Outside]]")
	}
	if !titles["AlsoOutside"] {
		t.Error("expected to find [[AlsoOutside]]")
	}
	if titles["Inside"] {
		t.Error("should NOT find [[Inside]] from fenced code block")
	}
	if len(links) != 2 {
		t.Errorf("expected 2 links, got %d: %v", len(links), links)
	}
}

func TestParseInlineTagsIgnoresFencedCode(t *testing.T) {
	text := "Normal #outside tag.\n```\n#inside should be ignored.\n```\nMore #alsooutside."
	masked := maskInertContent(text)
	tags := parseInlineTags(masked)

	tagSet := make(map[string]bool)
	for _, tag := range tags {
		tagSet[tag] = true
	}

	if !tagSet["outside"] {
		t.Error("expected to find #outside")
	}
	if !tagSet["alsooutside"] {
		t.Error("expected to find #alsooutside")
	}
	if tagSet["inside"] {
		t.Error("should NOT find #inside from fenced code block")
	}
}

func TestFindBacklinksIgnoresFencedCode(t *testing.T) {
	vaultDir := t.TempDir()

	// Note A links to B only inside a code block
	os.WriteFile(
		filepath.Join(vaultDir, "A.md"),
		[]byte("# A\n\nSome text.\n```\n[[B]] in code\n```\n"),
		0644,
	)

	// Note B exists
	os.WriteFile(
		filepath.Join(vaultDir, "B.md"),
		[]byte("# B\n\nContent.\n"),
		0644,
	)

	results, err := findBacklinks(vaultDir, "B")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 backlinks (link is inside code block), got %d: %v", len(results), results)
	}
}

func TestOrphansIgnoresFencedCode(t *testing.T) {
	vaultDir := t.TempDir()

	// A links to B ONLY inside a code block -- B should be orphaned
	os.WriteFile(
		filepath.Join(vaultDir, "A.md"),
		[]byte("# A\n\n```\n[[B]] in code\n```\n"),
		0644,
	)
	os.WriteFile(
		filepath.Join(vaultDir, "B.md"),
		[]byte("# B\n\nContent.\n"),
		0644,
	)

	// Capture orphans by examining the function behavior
	// cmdOrphans uses parseWikilinks which should now mask fenced content
	// B should appear as an orphan since the only link to it is inside a code block
	// We need to test the behavior through the public functions

	// Collect referenced titles the same way cmdOrphans does
	referenced := make(map[string]bool)
	filepath.WalkDir(vaultDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		for _, link := range parseWikilinks(string(data)) {
			referenced[strings.ToLower(link.Title)] = true
		}
		return nil
	})

	if referenced["b"] {
		t.Error("B should NOT be referenced (link is inside code block), so it should be an orphan")
	}
}

func TestUnresolvedIgnoresFencedCode(t *testing.T) {
	vaultDir := t.TempDir()

	// Note with [[Missing]] only inside a code block
	os.WriteFile(
		filepath.Join(vaultDir, "Source.md"),
		[]byte("# Source\n\n```\n[[Missing]] in code\n```\n"),
		0644,
	)

	// Simulate unresolved detection the same way cmdUnresolved does
	titles := make(map[string]bool)
	filepath.WalkDir(vaultDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		title := strings.TrimSuffix(d.Name(), ".md")
		titles[strings.ToLower(title)] = true
		return nil
	})

	var unresolved []string
	filepath.WalkDir(vaultDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		for _, link := range parseWikilinks(string(data)) {
			lower := strings.ToLower(link.Title)
			if !titles[lower] {
				unresolved = append(unresolved, link.Title)
			}
		}
		return nil
	})

	if len(unresolved) != 0 {
		t.Errorf("expected 0 unresolved links (link is inside code block), got %v", unresolved)
	}
}

func TestLinksIgnoresFencedCode(t *testing.T) {
	vaultDir := t.TempDir()

	// Note with [[Target]] both inside and outside code block
	os.WriteFile(
		filepath.Join(vaultDir, "Source.md"),
		[]byte("# Source\n\n```\n[[InsideOnly]] in code\n```\n[[Outside]] is real.\n"),
		0644,
	)
	os.WriteFile(
		filepath.Join(vaultDir, "Outside.md"),
		[]byte("# Outside\n"),
		0644,
	)

	// Read the source and parse links the same way cmdLinks does
	data, err := os.ReadFile(filepath.Join(vaultDir, "Source.md"))
	if err != nil {
		t.Fatal(err)
	}

	links := parseWikilinks(string(data))

	titles := make(map[string]bool)
	for _, l := range links {
		titles[l.Title] = true
	}

	if titles["InsideOnly"] {
		t.Error("should NOT find [[InsideOnly]] from fenced code block")
	}
	if !titles["Outside"] {
		t.Error("should find [[Outside]] from outside code block")
	}
}

func TestTagsIgnoresFencedCode(t *testing.T) {
	vaultDir := t.TempDir()

	os.WriteFile(
		filepath.Join(vaultDir, "note.md"),
		[]byte("# Note\n\n#real-tag\n\n```\n#hidden-tag\n```\n"),
		0644,
	)

	// Read and parse tags the same way cmdTags does through allNoteTags
	data, err := os.ReadFile(filepath.Join(vaultDir, "note.md"))
	if err != nil {
		t.Fatal(err)
	}

	tags := allNoteTags(string(data))
	tagSet := make(map[string]bool)
	for _, tag := range tags {
		tagSet[tag] = true
	}

	if !tagSet["real-tag"] {
		t.Error("should find #real-tag from outside code block")
	}
	if tagSet["hidden-tag"] {
		t.Error("should NOT find #hidden-tag from inside code block")
	}
}

func TestMermaidBlockIgnored(t *testing.T) {
	vaultDir := t.TempDir()

	os.WriteFile(
		filepath.Join(vaultDir, "diagram.md"),
		[]byte("# Diagram\n\n[[RealLink]]\n\n```mermaid\ngraph TD\nA[[FakeLink]] --> B\n```\n"),
		0644,
	)

	data, err := os.ReadFile(filepath.Join(vaultDir, "diagram.md"))
	if err != nil {
		t.Fatal(err)
	}

	links := parseWikilinks(string(data))
	titles := make(map[string]bool)
	for _, l := range links {
		titles[l.Title] = true
	}

	if !titles["RealLink"] {
		t.Error("should find [[RealLink]] outside mermaid block")
	}
	if titles["FakeLink"] {
		t.Error("should NOT find [[FakeLink]] inside mermaid block")
	}
}

// --- E2E Test ---

func TestE2EInertZoneFencedCode(t *testing.T) {
	vaultDir := t.TempDir()

	// Create a realistic vault with wikilinks and tags both inside and outside fenced code blocks

	// Note 1: Has both real and code-fenced links/tags
	os.WriteFile(
		filepath.Join(vaultDir, "Overview.md"),
		[]byte("---\ntags: [project]\n---\n\n# Overview\n\nSee [[Design Doc]] for details.\n\n#architecture\n\n```go\n// Example: [[FakeLink]] reference\nfmt.Println(\"#not-a-tag\")\n```\n\n```mermaid\ngraph TD\nA[[MermaidNode]] --> B\n```\n"),
		0644,
	)

	// Note 2: The real link target
	os.WriteFile(
		filepath.Join(vaultDir, "Design Doc.md"),
		[]byte("# Design Doc\n\nDetails here. See [[Overview]] for context.\n"),
		0644,
	)

	// Note 3: Only referenced inside a code block (should be orphaned)
	os.WriteFile(
		filepath.Join(vaultDir, "FakeLink.md"),
		[]byte("# FakeLink\n\nI should be an orphan because I'm only referenced in code blocks.\n"),
		0644,
	)

	// Note 4: Not referenced at all
	os.WriteFile(
		filepath.Join(vaultDir, "Island.md"),
		[]byte("# Island\n\nTruly unreferenced.\n"),
		0644,
	)

	// --- Test backlinks ---
	// "Design Doc" should have backlinks from "Overview" (real link)
	backlinks, err := findBacklinks(vaultDir, "Design Doc")
	if err != nil {
		t.Fatalf("findBacklinks Design Doc: %v", err)
	}
	if len(backlinks) != 1 || backlinks[0] != "Overview.md" {
		t.Errorf("Design Doc backlinks: got %v, want [Overview.md]", backlinks)
	}

	// "FakeLink" should have NO backlinks (only referenced in code block)
	backlinks, err = findBacklinks(vaultDir, "FakeLink")
	if err != nil {
		t.Fatalf("findBacklinks FakeLink: %v", err)
	}
	if len(backlinks) != 0 {
		t.Errorf("FakeLink should have 0 backlinks (code-only reference), got %v", backlinks)
	}

	// "MermaidNode" should have NO backlinks (only in mermaid block)
	backlinks, err = findBacklinks(vaultDir, "MermaidNode")
	if err != nil {
		t.Fatalf("findBacklinks MermaidNode: %v", err)
	}
	if len(backlinks) != 0 {
		t.Errorf("MermaidNode should have 0 backlinks, got %v", backlinks)
	}

	// --- Test links ---
	overviewData, _ := os.ReadFile(filepath.Join(vaultDir, "Overview.md"))
	links := parseWikilinks(string(overviewData))
	linkTitles := make(map[string]bool)
	for _, l := range links {
		linkTitles[l.Title] = true
	}

	if !linkTitles["Design Doc"] {
		t.Error("Overview should link to Design Doc")
	}
	if linkTitles["FakeLink"] {
		t.Error("Overview should NOT link to FakeLink (inside code block)")
	}
	if linkTitles["MermaidNode"] {
		t.Error("Overview should NOT link to MermaidNode (inside mermaid block)")
	}

	// --- Test orphans ---
	// Collect referenced titles (same logic as cmdOrphans)
	referenced := make(map[string]bool)
	filepath.WalkDir(vaultDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		for _, link := range parseWikilinks(string(data)) {
			referenced[strings.ToLower(link.Title)] = true
		}
		return nil
	})

	// FakeLink should be orphaned (only referenced in code block)
	if referenced["fakelink"] {
		t.Error("FakeLink should be unreferenced (code-only reference)")
	}
	// Island should be orphaned
	if referenced["island"] {
		t.Error("Island should be unreferenced")
	}
	// Design Doc should NOT be orphaned
	if !referenced["design doc"] {
		t.Error("Design Doc should be referenced by Overview")
	}
	// Overview should NOT be orphaned
	if !referenced["overview"] {
		t.Error("Overview should be referenced by Design Doc")
	}

	// --- Test unresolved ---
	// Collect all note titles
	titles := make(map[string]bool)
	filepath.WalkDir(vaultDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		title := strings.TrimSuffix(d.Name(), ".md")
		titles[strings.ToLower(title)] = true
		return nil
	})

	// Collect unresolved links
	var unresolvedLinks []string
	filepath.WalkDir(vaultDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		for _, link := range parseWikilinks(string(data)) {
			lower := strings.ToLower(link.Title)
			if !titles[lower] {
				unresolvedLinks = append(unresolvedLinks, link.Title)
			}
		}
		return nil
	})

	// FakeLink exists as a note, so even if it were linked, it wouldn't be unresolved.
	// MermaidNode does NOT exist as a note, but the link is inside a code block,
	// so it should NOT appear as unresolved.
	for _, u := range unresolvedLinks {
		if u == "MermaidNode" {
			t.Error("MermaidNode should NOT be unresolved (link is inside code block)")
		}
		if u == "FakeLink" {
			t.Error("FakeLink should NOT be unresolved (link is inside code block)")
		}
	}

	// --- Test tags ---
	overviewTags := allNoteTags(string(overviewData))
	tagSet := make(map[string]bool)
	for _, tag := range overviewTags {
		tagSet[tag] = true
	}

	// Frontmatter tag should be found
	if !tagSet["project"] {
		t.Error("should find frontmatter tag 'project'")
	}
	// Inline tag outside code should be found
	if !tagSet["architecture"] {
		t.Error("should find inline tag 'architecture'")
	}
	// Tag inside code block should NOT be found
	if tagSet["not-a-tag"] {
		t.Error("should NOT find 'not-a-tag' from inside code block")
	}

	// --- Verify orphans are correctly sorted ---
	type noteInfo struct {
		relPath string
		title   string
	}
	var notes []noteInfo
	filepath.WalkDir(vaultDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		title := strings.TrimSuffix(d.Name(), ".md")
		relPath, _ := filepath.Rel(vaultDir, path)
		notes = append(notes, noteInfo{relPath: relPath, title: title})
		return nil
	})

	var orphans []string
	for _, note := range notes {
		if !referenced[strings.ToLower(note.title)] {
			orphans = append(orphans, note.relPath)
		}
	}
	sort.Strings(orphans)

	// FakeLink.md and Island.md should be orphans
	expectedOrphans := []string{"FakeLink.md", "Island.md"}
	if len(orphans) != len(expectedOrphans) {
		t.Errorf("expected %d orphans, got %d: %v", len(expectedOrphans), len(orphans), orphans)
	} else {
		for i, want := range expectedOrphans {
			if orphans[i] != want {
				t.Errorf("orphan[%d] = %q, want %q", i, orphans[i], want)
			}
		}
	}
}
