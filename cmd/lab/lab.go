package lab

import (
	"fmt"
	"os"

	prompt "github.com/c-bata/go-prompt"
	"github.com/chazzy/g2o/internal/glclient"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "lab",
	Short: "Interact with GitLab",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, ok := os.LookupEnv("GITLAB_TOKEN")
		if !ok {
			return fmt.Errorf("GITLAB_TOKEN is not set")
		}

		g, err := glclient.NewGitlab(token)
		if err != nil {
			return err
		}

		return runREPL(g)
	},
}

func runREPL(g glclient.GitLab) error {
	fmt.Println("GitLab REPL — type 'exit' to quit, 'help' for commands.")

	executor := func(in string) {
		switch in {
		case "exit", "quit":
			os.Exit(0)
		case "help":
			fmt.Println("  groups   list your groups")
			fmt.Println("  projects list your projects")
			fmt.Println("  me       show current user")
			fmt.Println("  exit     quit")
		case "groups":
			if err := g.RunGroups(); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		case "projects":
			if err := g.RunProjects(); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		case "me":
			if err := g.RunCurrentUser(); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		case "":
		default:
			fmt.Printf("unknown command: %q — type 'help' for available commands\n", in)
		}
	}

	completer := func(d prompt.Document) []prompt.Suggest {
		suggests := []prompt.Suggest{
			{Text: "groups", Description: "List your groups"},
			{Text: "projects", Description: "List your projects"},
			{Text: "me", Description: "Show current user"},
			{Text: "exit", Description: "Quit"},
		}
		return prompt.FilterHasPrefix(suggests, d.GetWordBeforeCursor(), true)
	}

	prompt.New(executor, completer,
		prompt.OptionPrefix("g2o> "),
		prompt.OptionTitle("g2o GitLab REPL"),
	).Run()

	return nil
}
