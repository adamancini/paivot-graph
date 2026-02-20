package main

import "regexp"

// maskPass is a function that masks one type of inert zone.
// Each pass receives the text (potentially already partially masked by
// earlier passes) and returns the text with its zone type masked.
type maskPass func(text string) string

// inertPasses is the ordered slice of mask functions.
// Each story adds its pass to this slice via an init() function or by
// calling registerMaskPass. Order matters:
// fenced code blocks first, then inline code, then comments, then math.
var inertPasses []maskPass

// registerMaskPass adds a masking pass. Called during init.
func registerMaskPass(p maskPass) {
	inertPasses = append(inertPasses, p)
}

// maskInertContent applies all registered masking passes in order.
// The result has the same byte length and line count as the input,
// but content inside inert zones is replaced with spaces (preserving newlines).
func maskInertContent(text string) string {
	for _, pass := range inertPasses {
		text = pass(text)
	}
	return text
}

// maskRegion replaces all non-newline characters in text[start:end] with spaces.
// Newlines are preserved so that line numbers remain stable.
func maskRegion(text []byte, start, end int) {
	for i := start; i < end; i++ {
		if text[i] != '\n' {
			text[i] = ' '
		}
	}
}

// fencedCodePattern matches the opening fence of a fenced code block:
// three or more backticks at the start of a line, optionally followed by a
// language identifier, then a newline.
var fencedCodePattern = regexp.MustCompile("(?m)^(```\\w*)\n")

// closingFencePattern matches a closing fence: three backticks at the start
// of a line (possibly followed by whitespace and then end-of-line or end-of-string).
var closingFencePattern = regexp.MustCompile("(?m)^```[ \t]*$")

// maskFencedCodeBlocks masks the content inside fenced code blocks (``` ... ```).
// The fence delimiters themselves are NOT masked.
// Unclosed fences at EOF: mask to end of file (matches Obsidian behavior).
func maskFencedCodeBlocks(text string) string {
	buf := []byte(text)
	pos := 0

	for pos < len(buf) {
		// Find the next opening fence
		loc := fencedCodePattern.FindIndex(buf[pos:])
		if loc == nil {
			break
		}

		// The content to mask starts after the opening fence line (after the \n)
		openEnd := pos + loc[1] // position right after the opening fence line's newline
		contentStart := openEnd

		// Find the closing fence starting from content area
		closeLoc := closingFencePattern.FindIndex(buf[contentStart:])
		if closeLoc == nil {
			// Unclosed fence: mask to end of file
			maskRegion(buf, contentStart, len(buf))
			break
		}

		// Mask from content start to the start of the closing fence line
		contentEnd := contentStart + closeLoc[0]
		maskRegion(buf, contentStart, contentEnd)

		// Move past the closing fence
		pos = contentStart + closeLoc[1]
	}

	return string(buf)
}

func init() {
	registerMaskPass(maskFencedCodeBlocks)
}
