package lab

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	prompt "github.com/c-bata/go-prompt"
	"github.com/chazzychouse/g2o/internal/glclient"
	"github.com/chazzychouse/g2o/internal/store"
	"github.com/chazzychouse/g2o/internal/styles"
	gosync "github.com/chazzychouse/g2o/internal/sync"
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

		// Open local store.
		db, err := store.Open(store.DefaultPath())
		if err != nil {
			return fmt.Errorf("open store: %w", err)
		}
		defer db.Close()

		var opts []glclient.Option
		opts = append(opts, glclient.WithStore(db))

		if _, ok := os.LookupEnv("G2O_DEBUG"); ok {
			logFile, err := os.Create("g2o.debug.log")
			if err != nil {
				return fmt.Errorf("failed to create debug log: %w", err)
			}
			defer logFile.Close()
			logger := slog.New(slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug}))
			opts = append(opts, glclient.WithLogger(logger))

			dumpFile, err := os.Create("g2o.issues.jsonl")
			if err != nil {
				return fmt.Errorf("failed to create dump file: %w", err)
			}
			defer dumpFile.Close()
			opts = append(opts, glclient.WithDump(dumpFile))
		}

		g, err := glclient.NewGitlab(token, opts...)
		if err != nil {
			return err
		}

		syncer := gosync.NewSyncer(&g, db)

		// Auto-sync on first run if the database is empty.
		if syncer.NeedsFullSync() {
			fmt.Println(styles.Title.Render("First run detected — syncing data from GitLab..."))
			if err := syncer.SyncAll(context.Background()); err != nil {
				fmt.Fprintln(os.Stderr, styles.Error.Render("sync failed: "+err.Error()))
				fmt.Println("Continuing with API fallback...")
			}
		}

		return runREPL(g, syncer)
	},
}

func runREPL(g glclient.GitLab, syncer *gosync.Syncer) error {
	var cmds []*replCmd
	cmds = []*replCmd{
		{Name: "groups", Desc: "List your groups", Run: func(args []string) error { return g.RunGroups() }},
		{
			Name: "group", Desc: "Group commands", Arg: "<gid>",
			Sub: []*replCmd{
				{Name: "issues", Desc: "List issues for group", Run: func(args []string) error { return g.RunGroupsIssues(context.Background(), args[0]) }},
			},
		},
		{Name: "projects", Desc: "List your projects", Run: func(args []string) error { return g.RunProjects() }},
		{Name: "me", Desc: "Show current user", Run: func(args []string) error { return g.RunCurrentUser() }},
		{Name: "issues", Desc: "List your issues", Run: func(args []string) error { return g.RunIssues() }},
		{
			Name: "sync", Desc: "Sync data from GitLab",
			Run: func(args []string) error { return syncer.SyncIncremental(context.Background()) },
			Sub: []*replCmd{
				{Name: "full", Desc: "Full re-download of all data", Run: func(args []string) error { return syncer.SyncAll(context.Background()) }},
				{
					Name: "groups", Desc: "Sync groups only",
					Run: func(args []string) error { return syncer.SyncGroups(context.Background()) },
					Sub: []*replCmd{
						{Name: "issues", Desc: "Sync issues for all groups", Run: func(args []string) error { return syncer.SyncGroupIssues(context.Background()) }},
					},
				},
				{Name: "projects", Desc: "Sync projects only", Run: func(args []string) error { return syncer.SyncProjects(context.Background()) }},
				{Name: "issues", Desc: "Sync issues only", Run: func(args []string) error { return syncer.SyncIssues(context.Background()) }},
				{Name: "status", Desc: "Show sync timestamps", Run: func(args []string) error { return syncer.ShowStatus() }},
			},
		},
		{Name: "help", Desc: "Show help", Run: func(args []string) error { buildHelp(cmds); return nil }},
		{Name: "exit", Desc: "Quit"},
		{Name: "quit", Desc: "Quit"},
	}

	fmt.Println(styles.Banner.Render("GitLab REPL") + " — type 'exit' to quit, 'help' for commands.")
	buildHelp(cmds)

	executor := func(in string) {
		parts := strings.Fields(in)
		if len(parts) == 0 {
			return
		}
		if parts[0] == "exit" || parts[0] == "quit" {
			return // handled by OptionSetExitCheckerOnInput
		}
		if err := dispatch(cmds, parts); err != nil {
			fmt.Fprintln(os.Stderr, styles.Error.Render(err.Error()))
		}
	}

	var showAll bool

	completer := func(d prompt.Document) []prompt.Suggest {
		text := d.TextBeforeCursor()
		tokens := strings.Fields(text)

		// If the cursor is right after a space, the current word is empty
		// but we still want context-aware suggestions for the next position.
		word := d.GetWordBeforeCursor()
		if word == "" && text == "" {
			if showAll {
				showAll = false
				return complete(cmds, nil)
			}
			return nil
		}

		// Determine which tokens are "completed" (before the word being typed).
		completed := tokens
		if word != "" && len(completed) > 0 {
			completed = completed[:len(completed)-1]
		}

		suggestions := complete(cmds, completed)
		if word == "" {
			return suggestions
		}
		return prompt.FilterHasPrefix(suggestions, word, true)
	}

	defer fmt.Print("\033[?25h\033[0m\r\n") // restore cursor, reset attrs on exit

	prompt.New(executor, completer,
		prompt.OptionPrefix("❯ "),
		prompt.OptionTitle("g2o GitLab REPL"),
		prompt.OptionSetExitCheckerOnInput(func(in string, breakline bool) bool {
			parts := strings.Fields(in)
			return len(parts) > 0 && (parts[0] == "exit" || parts[0] == "quit")
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
