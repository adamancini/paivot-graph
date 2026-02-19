// vlt -- fast Obsidian vault CLI (no app required)
//
// Drop-in replacement for the obsidian CLI that operates directly on the
// filesystem. No Obsidian app dependency, no Electron round-trips.
//
// Discovers vaults from the Obsidian config file, resolves notes by title,
// and performs read/search/create/append/move/property:set operations.
package main

import (
	"fmt"
	"os"
	"strings"
)

const version = "0.2.0"

var knownCommands = map[string]bool{
	"read": true, "search": true, "create": true,
	"append": true, "move": true, "property:set": true,
	"backlinks": true, "links": true,
	"vaults": true, "help": true, "version": true,
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd, params, flags := parseArgs(os.Args[1:])

	if cmd == "help" || flags["--help"] || flags["-h"] {
		usage()
		return
	}
	if cmd == "version" || flags["--version"] {
		fmt.Println("vlt " + version)
		return
	}
	if cmd == "vaults" {
		if err := cmdVaults(); err != nil {
			die("%v", err)
		}
		return
	}
	if cmd == "" {
		die("no command specified. Run 'vlt help' for usage.")
	}

	// Resolve vault
	vaultName := params["vault"]
	if vaultName == "" {
		vaultName = os.Getenv("VLT_VAULT")
	}
	if vaultName == "" {
		die("vault not specified. Use vault=\"<name>\" or set VLT_VAULT env var.")
	}

	vaultDir, err := resolveVault(vaultName)
	if err != nil {
		die("%v", err)
	}

	// Dispatch
	switch cmd {
	case "read":
		err = cmdRead(vaultDir, params)
	case "search":
		err = cmdSearch(vaultDir, params)
	case "create":
		err = cmdCreate(vaultDir, params, flags["silent"])
	case "append":
		err = cmdAppend(vaultDir, params)
	case "move":
		err = cmdMove(vaultDir, params)
	case "property:set":
		err = cmdPropertySet(vaultDir, params)
	case "backlinks":
		err = cmdBacklinks(vaultDir, params)
	case "links":
		err = cmdLinks(vaultDir, params)
	default:
		die("unknown command: %s", cmd)
	}

	if err != nil {
		die("%v", err)
	}
}

// parseArgs splits CLI arguments into a command name, key=value parameters,
// and bare-word flags. It preserves the obsidian CLI's key="value" syntax.
func parseArgs(args []string) (string, map[string]string, map[string]bool) {
	params := make(map[string]string)
	flags := make(map[string]bool)
	var cmd string

	for _, arg := range args {
		if i := strings.Index(arg, "="); i > 0 {
			key := arg[:i]
			val := arg[i+1:]
			// Strip surrounding quotes (shouldn't be needed after shell parsing,
			// but handles edge cases like programmatic invocation).
			val = strings.Trim(val, "\"'")
			params[key] = val
		} else if knownCommands[arg] {
			cmd = arg
		} else {
			flags[arg] = true
		}
	}

	return cmd, params, flags
}

func die(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "vlt: "+format+"\n", args...)
	os.Exit(1)
}

func usage() {
	fmt.Print(`vlt -- fast Obsidian vault CLI (no app required)

Usage:
  vlt vault="<name>" <command> [args...]

Commands:
  read         file="<title>"                            Read a note by title
  search       query="<term>"                            Search notes
  create       name="<title>" path="<path>" [content="<text>"] [silent]
                                                         Create a note
  append       file="<title>" [content="<text>"]         Append to a note
  move         path="<from>" to="<to>"                   Move/rename (updates wikilinks)
  property:set file="<title>" name="<key>" value="<val>" Set frontmatter property
  backlinks    file="<title>"                            Find notes linking to this note
  links        file="<title>"                            List outgoing links (flags broken)
  vaults                                                 List discovered vaults

Options:
  vault="<name>"   Vault name (from Obsidian config) or absolute path.
                   Also settable via VLT_VAULT env var.
  silent           Suppress output on create.

Content from stdin:
  If content= is omitted for create/append, content is read from stdin.
  This avoids shell argument-length limits for large notes.

Examples:
  vlt vault="Claude" read file="Session Operating Mode"
  vlt vault="Claude" search query="paivot"
  vlt vault="Claude" create name="My Note" path="_inbox/My Note.md" content="# Hello" silent
  echo "## Update" | vlt vault="Claude" append file="My Note"
  vlt vault="Claude" move path="_inbox/Note.md" to="decisions/Note.md"
  vlt vault="Claude" move path="_inbox/Old Name.md" to="decisions/New Name.md"  # updates all [[Old Name]] refs
  vlt vault="Claude" property:set file="Note" name="status" value="archived"
  vlt vault="Claude" backlinks file="Session Operating Mode"
  vlt vault="Claude" links file="Developer Agent"
  vlt vaults
`)
}
