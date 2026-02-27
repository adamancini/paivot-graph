package lifecycle

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/RamXX/paivot-graph/pvg-cli/internal/dispatcher"
)

// userPromptInput matches the JSON Claude Code sends to UserPromptSubmit hooks.
type userPromptInput struct {
	Prompt string `json:"prompt"`
}

// triggerPhrases are case-insensitive phrases that activate dispatcher mode.
var triggerPhrases = []string{
	"use paivot",
	"paivot this",
	"run paivot",
	"engage paivot",
	"with paivot",
}

// UserPromptSubmit detects Paivot trigger phrases in user prompts and
// auto-enables dispatcher mode. Outputs JSON with additionalContext when
// dispatcher mode is activated.
func UserPromptSubmit() error {
	var input userPromptInput
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		return nil // fail-open
	}

	if !containsTriggerPhrase(input.Prompt) {
		return nil // silent exit
	}

	cwd, _ := os.Getwd()
	if cwd == "" {
		return nil
	}

	// Enable dispatcher mode
	if err := dispatcher.On(cwd); err != nil {
		// Log but don't block
		fmt.Fprintf(os.Stderr, "pvg: failed to enable dispatcher mode: %v\n", err)
		return nil
	}

	// Output hook response with context reinforcement
	resp := map[string]any{
		"hookSpecificOutput": map[string]any{
			"hookEventName":     "UserPromptSubmit",
			"additionalContext": "DISPATCHER MODE ACTIVE. You are a coordinator only. Do NOT write D&F files, source code, or stories directly. Spawn the appropriate agent instead.",
		},
	}
	return json.NewEncoder(os.Stdout).Encode(resp)
}

// containsTriggerPhrase checks if the prompt contains any Paivot trigger phrase.
func containsTriggerPhrase(prompt string) bool {
	lower := strings.ToLower(prompt)
	for _, phrase := range triggerPhrases {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}
