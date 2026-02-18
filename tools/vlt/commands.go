package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// cmdVaults lists all Obsidian vaults discovered from the config file.
func cmdVaults() error {
	vaults, err := discoverVaults()
	if err != nil {
		return err
	}

	if len(vaults) == 0 {
		fmt.Println("No vaults found.")
		return nil
	}

	// Sort by name for stable output
	names := make([]string, 0, len(vaults))
	for name := range vaults {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		fmt.Printf("%s\t%s\n", name, vaults[name])
	}
	return nil
}

// cmdRead prints the contents of a note resolved by title.
func cmdRead(vaultDir string, params map[string]string) error {
	title := params["file"]
	if title == "" {
		return fmt.Errorf("read requires file=\"<title>\"")
	}

	path, err := resolveNote(vaultDir, title)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	fmt.Print(string(data))
	return nil
}

// cmdSearch finds notes whose title or content matches the query (case-insensitive).
func cmdSearch(vaultDir string, params map[string]string) error {
	query := params["query"]
	if query == "" {
		return fmt.Errorf("search requires query=\"<term>\"")
	}

	queryLower := strings.ToLower(query)
	pathFilter := params["path"] // optional: limit to a subdirectory

	searchRoot := vaultDir
	if pathFilter != "" {
		searchRoot = filepath.Join(vaultDir, pathFilter)
		if _, err := os.Stat(searchRoot); os.IsNotExist(err) {
			return fmt.Errorf("path filter %q not found in vault", pathFilter)
		}
	}

	type result struct {
		title   string
		relPath string
	}
	var results []result

	err := filepath.WalkDir(searchRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		name := d.Name()
		if d.IsDir() && (strings.HasPrefix(name, ".") || name == ".trash") {
			return filepath.SkipDir
		}

		if d.IsDir() || !strings.HasSuffix(name, ".md") {
			return nil
		}

		title := strings.TrimSuffix(name, ".md")
		relPath, _ := filepath.Rel(vaultDir, path)

		// Check title first (cheap)
		if strings.Contains(strings.ToLower(title), queryLower) {
			results = append(results, result{title, relPath})
			return nil
		}

		// Check content (needs I/O)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		if strings.Contains(strings.ToLower(string(data)), queryLower) {
			results = append(results, result{title, relPath})
		}

		return nil
	})

	if err != nil {
		return err
	}

	if len(results) == 0 {
		return nil // silent on no results, matching grep convention
	}

	for _, r := range results {
		fmt.Printf("%s (%s)\n", r.title, r.relPath)
	}
	return nil
}

// cmdCreate creates a new note at the given path within the vault.
// Content comes from the content= parameter or stdin.
func cmdCreate(vaultDir string, params map[string]string, silent bool) error {
	name := params["name"]
	notePath := params["path"]

	if name == "" || notePath == "" {
		return fmt.Errorf("create requires name=\"<title>\" path=\"<relative-path>\"")
	}

	fullPath := filepath.Join(vaultDir, notePath)

	// Don't overwrite existing notes
	if _, err := os.Stat(fullPath); err == nil {
		if !silent {
			fmt.Fprintf(os.Stderr, "note already exists: %s\n", notePath)
		}
		return nil
	}

	content := params["content"]
	if content == "" {
		content = readStdinIfPiped()
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return err
	}

	if !silent {
		fmt.Printf("created: %s\n", notePath)
	}
	return nil
}

// cmdAppend adds content to the end of an existing note.
// Content comes from the content= parameter or stdin.
func cmdAppend(vaultDir string, params map[string]string) error {
	title := params["file"]
	if title == "" {
		return fmt.Errorf("append requires file=\"<title>\"")
	}

	path, err := resolveNote(vaultDir, title)
	if err != nil {
		return err
	}

	content := params["content"]
	if content == "" {
		content = readStdinIfPiped()
	}
	if content == "" {
		return fmt.Errorf("no content provided (use content=\"...\" or pipe to stdin)")
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprint(f, content)
	return err
}

// cmdMove moves a note from one path to another within the vault.
func cmdMove(vaultDir string, params map[string]string) error {
	from := params["path"]
	to := params["to"]

	if from == "" || to == "" {
		return fmt.Errorf("move requires path=\"<from>\" to=\"<to>\"")
	}

	fromPath := filepath.Join(vaultDir, from)
	toPath := filepath.Join(vaultDir, to)

	if _, err := os.Stat(fromPath); os.IsNotExist(err) {
		return fmt.Errorf("source not found: %s", from)
	}

	if err := os.MkdirAll(filepath.Dir(toPath), 0755); err != nil {
		return err
	}

	if err := os.Rename(fromPath, toPath); err != nil {
		return err
	}

	fmt.Printf("moved: %s -> %s\n", from, to)
	return nil
}

// cmdPropertySet sets or adds a YAML frontmatter property in a note.
func cmdPropertySet(vaultDir string, params map[string]string) error {
	title := params["file"]
	propName := params["name"]
	propValue := params["value"]

	if title == "" || propName == "" {
		return fmt.Errorf("property:set requires file=\"<title>\" name=\"<key>\" value=\"<val>\"")
	}

	path, err := resolveNote(vaultDir, title)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")

	// Find frontmatter boundaries (--- ... ---)
	fmStart, fmEnd := -1, -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			if fmStart == -1 {
				fmStart = i
			} else {
				fmEnd = i
				break
			}
		}
	}

	if fmStart == -1 || fmEnd == -1 {
		return fmt.Errorf("no frontmatter found in %q", title)
	}

	// Look for existing property line
	found := false
	prefix := propName + ":"
	for i := fmStart + 1; i < fmEnd; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, prefix) {
			lines[i] = fmt.Sprintf("%s: %s", propName, propValue)
			found = true
			break
		}
	}

	// If not found, insert before closing ---
	if !found {
		newLine := fmt.Sprintf("%s: %s", propName, propValue)
		// Insert newLine at position fmEnd
		lines = append(lines[:fmEnd+1], lines[fmEnd:]...)
		lines[fmEnd] = newLine
	}

	result := strings.Join(lines, "\n")
	if err := os.WriteFile(path, []byte(result), 0644); err != nil {
		return err
	}

	fmt.Printf("set %s=%s in %q\n", propName, propValue, title)
	return nil
}

// readStdinIfPiped reads all of stdin if it's being piped (not a terminal).
// Returns empty string if stdin is a terminal.
func readStdinIfPiped() string {
	stat, _ := os.Stdin.Stat()
	if stat.Mode()&os.ModeCharDevice != 0 {
		return "" // stdin is a terminal, not piped
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return ""
	}
	return string(data)
}
