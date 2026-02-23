package lab

import (
	"fmt"
	"strings"

	prompt "github.com/c-bata/go-prompt"
	"github.com/chazzychouse/g2o/internal/styles"
)

// replCmd is a node in the command tree. Each node matches a literal token
// (Name) or, when Arg is set, captures the next token as a positional argument.
type replCmd struct {
	Name string              // literal token to match (e.g. "group")
	Desc string              // shown in help and completion
	Arg  string              // if non-empty, next token is captured (e.g. "<gid>")
	Run  func(args []string) error // executor; args contains captured positional values
	Sub  []*replCmd          // subcommands
}

// dispatch walks the command tree and invokes the deepest matching Run.
// Positional args captured via Arg fields are collected in order.
func dispatch(cmds []*replCmd, tokens []string) error {
	var args []string
	nodes := cmds

	var matched *replCmd
	i := 0
	for i < len(tokens) {
		tok := tokens[i]
		found := false
		for _, c := range nodes {
			if c.Name == tok {
				matched = c
				i++
				// If this node expects a positional arg, consume the next token.
				if c.Arg != "" && i < len(tokens) {
					args = append(args, tokens[i])
					i++
				}
				nodes = c.Sub
				found = true
				break
			}
		}
		if !found {
			break
		}
	}

	if matched == nil || matched.Run == nil {
		return fmt.Errorf("unknown command: %q", strings.Join(tokens, " "))
	}
	return matched.Run(args)
}

// complete walks the command tree following already-typed tokens and returns
// suggestions for the next position.
func complete(cmds []*replCmd, tokens []string) []prompt.Suggest {
	nodes := cmds
	i := 0
	for i < len(tokens) {
		tok := tokens[i]
		found := false
		for _, c := range nodes {
			if c.Name == tok {
				i++
				// Skip over a positional arg value if present.
				if c.Arg != "" {
					if i < len(tokens) {
						i++
					} else {
						// User hasn't typed the arg yet — no suggestions.
						return nil
					}
				}
				nodes = c.Sub
				found = true
				break
			}
		}
		if !found {
			break
		}
	}

	var out []prompt.Suggest
	for _, c := range nodes {
		out = append(out, prompt.Suggest{Text: c.Name, Description: c.Desc})
	}
	return out
}

// buildHelp prints the command tree as a flat help listing.
func buildHelp(cmds []*replCmd) {
	printTree(cmds, "")
}

func printTree(cmds []*replCmd, prefix string) {
	for _, c := range cmds {
		label := prefix + c.Name
		if c.Arg != "" {
			label += " " + c.Arg
		}
		if c.Run != nil {
			// Pad to a fixed width for alignment.
			padded := label
			if len(padded) < 20 {
				padded += strings.Repeat(" ", 20-len(padded))
			}
			fmt.Printf("  %s  %s\n", styles.HelpCmd.Render(padded), styles.HelpDesc.Render(c.Desc))
		}
		if len(c.Sub) > 0 {
			printTree(c.Sub, label+" ")
		}
	}
}
