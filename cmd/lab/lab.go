package lab

import (
	"fmt"
	"os"

	prompt "github.com/c-bata/go-prompt"
	"github.com/chazzy/g2o/internal/glclient"
	"github.com/chazzy/g2o/internal/styles"
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
	fmt.Println(styles.Banner.Render("GitLab REPL") + " — type 'exit' to quit, 'help' for commands.")
	printHelp()
	executor := func(in string) {
		switch in {
		case "exit", "quit":
			// handled by OptionSetExitCheckerOnInput
		case "help":
			printHelp()
		case "groups":
			if err := g.RunGroups(); err != nil {
				fmt.Fprintln(os.Stderr, styles.Error.Render(err.Error()))
			}
		case "projects":
			if err := g.RunProjects(); err != nil {
				fmt.Fprintln(os.Stderr, styles.Error.Render(err.Error()))
			}
		case "me":
			if err := g.RunCurrentUser(); err != nil {
				fmt.Fprintln(os.Stderr, styles.Error.Render(err.Error()))
			}
		case "":
		default:
			fmt.Println(styles.Error.Render(fmt.Sprintf("unknown command: %q", in)) + " — type 'help' for available commands")
		}
	}

	suggests := []prompt.Suggest{
		{Text: "groups", Description: "List your groups"},
		{Text: "projects", Description: "List your projects"},
		{Text: "me", Description: "Show current user"},
		{Text: "help", Description: "Show help"},
		{Text: "exit", Description: "Quit"},
	}

	var showAll bool

	completer := func(d prompt.Document) []prompt.Suggest {
		word := d.GetWordBeforeCursor()
		if word == "" {
			if showAll {
				showAll = false
				return suggests
			}
			return nil
		}
		return prompt.FilterHasPrefix(suggests, word, true)
	}

	defer fmt.Print("\033[?25h\033[0m\r\n") // restore cursor, reset attrs on exit

	prompt.New(executor, completer,
		prompt.OptionPrefix("❯ "),
		prompt.OptionTitle("g2o GitLab REPL"),
		prompt.OptionSetExitCheckerOnInput(func(in string, breakline bool) bool {
			return in == "exit" || in == "quit"
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlSpace,
			Fn: func(buf *prompt.Buffer) {
				showAll = true
			},
		}),
	).Run()

	return nil
}

func printHelp() {
	fmt.Printf("  %s  %s\n", styles.HelpCmd.Render("groups  "), styles.HelpDesc.Render("list your groups"))
	fmt.Printf("  %s  %s\n", styles.HelpCmd.Render("projects"), styles.HelpDesc.Render("list your projects"))
	fmt.Printf("  %s  %s\n", styles.HelpCmd.Render("me      "), styles.HelpDesc.Render("show current user"))
	fmt.Printf("  %s  %s\n", styles.HelpCmd.Render("exit    "), styles.HelpDesc.Render("quit"))
}
