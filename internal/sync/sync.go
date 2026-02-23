package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/chazzychouse/g2o/internal/glclient"
	"github.com/chazzychouse/g2o/internal/store"
	"github.com/chazzychouse/g2o/internal/styles"
)

type Syncer struct {
	client *glclient.GitLab
	store  *store.Store
}

func NewSyncer(client *glclient.GitLab, store *store.Store) *Syncer {
	return &Syncer{client: client, store: store}
}

// SyncAll performs a full sync of all resources, downloading everything
// and removing stale records.
func (s *Syncer) SyncAll(ctx context.Context) error {
	fmt.Println(styles.Title.Render("Full sync starting..."))
	now := time.Now().UTC()

	if err := s.syncUser(ctx); err != nil {
		return fmt.Errorf("sync user: %w", err)
	}
	if err := s.syncGroupsFull(ctx); err != nil {
		return fmt.Errorf("sync groups: %w", err)
	}
	if err := s.syncProjectsFull(ctx); err != nil {
		return fmt.Errorf("sync projects: %w", err)
	}
	if err := s.syncIssuesFull(ctx); err != nil {
		return fmt.Errorf("sync issues: %w", err)
	}

	// Mark full sync timestamps.
	for _, res := range []string{"groups", "projects", "issues", "user"} {
		if err := s.store.SetFullSync(res, now); err != nil {
			return fmt.Errorf("set full sync %s: %w", res, err)
		}
	}

	fmt.Println(styles.Success.Render("Full sync complete."))
	return nil
}

// SyncIncremental uses timestamps from last sync for incremental updates.
func (s *Syncer) SyncIncremental(ctx context.Context) error {
	fmt.Println(styles.Title.Render("Incremental sync starting..."))
	now := time.Now().UTC()

	if err := s.syncUser(ctx); err != nil {
		return fmt.Errorf("sync user: %w", err)
	}
	// Groups always full (no UpdatedAfter on API).
	if err := s.syncGroupsFull(ctx); err != nil {
		return fmt.Errorf("sync groups: %w", err)
	}
	if err := s.syncProjectsIncremental(ctx); err != nil {
		return fmt.Errorf("sync projects: %w", err)
	}
	if err := s.syncIssuesIncremental(ctx); err != nil {
		return fmt.Errorf("sync issues: %w", err)
	}

	for _, res := range []string{"groups", "projects", "issues", "user"} {
		if err := s.store.SetLastSynced(res, now); err != nil {
			return fmt.Errorf("set last synced %s: %w", res, err)
		}
	}

	fmt.Println(styles.Success.Render("Sync complete."))
	return nil
}

// SyncGroups syncs groups only.
func (s *Syncer) SyncGroups(ctx context.Context) error {
	if err := s.syncGroupsFull(ctx); err != nil {
		return err
	}
	return s.store.SetLastSynced("groups", time.Now().UTC())
}

// SyncProjects syncs projects only.
func (s *Syncer) SyncProjects(ctx context.Context) error {
	if err := s.syncProjectsIncremental(ctx); err != nil {
		return err
	}
	return s.store.SetLastSynced("projects", time.Now().UTC())
}

// SyncIssues syncs issues only.
func (s *Syncer) SyncIssues(ctx context.Context) error {
	if err := s.syncIssuesIncremental(ctx); err != nil {
		return err
	}
	return s.store.SetLastSynced("issues", time.Now().UTC())
}

func (s *Syncer) syncUser(ctx context.Context) error {
	fmt.Print("  syncing user... ")
	u, err := s.client.CurrentUser()
	if err != nil {
		return err
	}
	if err := s.store.UpsertUser(convertUser(u)); err != nil {
		return err
	}
	fmt.Println(styles.Success.Render("done"))
	return nil
}

func (s *Syncer) syncGroupsFull(ctx context.Context) error {
	fmt.Print("  syncing groups... ")
	groups, err := s.client.AllGroups(ctx)
	if err != nil {
		return err
	}
	sg := convertGroups(groups)
	if err := s.store.UpsertGroups(sg); err != nil {
		return err
	}
	// Delete stale.
	ids := make([]int64, len(groups))
	for i, g := range groups {
		ids[i] = g.ID
	}
	if err := s.store.DeleteStaleGroups(ids); err != nil {
		return err
	}
	fmt.Println(styles.Success.Render(fmt.Sprintf("%d groups", len(groups))))
	return nil
}

func (s *Syncer) syncProjectsFull(ctx context.Context) error {
	fmt.Print("  syncing projects... ")
	projects, err := s.client.AllProjects(ctx, nil)
	if err != nil {
		return err
	}
	sp := convertProjects(projects)
	if err := s.store.UpsertProjects(sp); err != nil {
		return err
	}
	ids := make([]int64, len(projects))
	for i, p := range projects {
		ids[i] = p.ID
	}
	if err := s.store.DeleteStaleProjects(ids); err != nil {
		return err
	}
	fmt.Println(styles.Success.Render(fmt.Sprintf("%d projects", len(projects))))
	return nil
}

func (s *Syncer) syncProjectsIncremental(ctx context.Context) error {
	fmt.Print("  syncing projects... ")
	lastSynced, err := s.store.GetLastSynced("projects")
	if err != nil {
		return err
	}
	var after *time.Time
	if !lastSynced.IsZero() {
		after = &lastSynced
	}
	projects, err := s.client.AllProjects(ctx, after)
	if err != nil {
		return err
	}
	if len(projects) > 0 {
		sp := convertProjects(projects)
		if err := s.store.UpsertProjects(sp); err != nil {
			return err
		}
	}
	fmt.Println(styles.Success.Render(fmt.Sprintf("%d projects", len(projects))))
	return nil
}

func (s *Syncer) syncIssuesFull(ctx context.Context) error {
	fmt.Print("  syncing issues... ")
	issues, err := s.client.AllIssues(ctx, nil)
	if err != nil {
		return err
	}
	si := convertIssues(issues)
	if err := s.store.UpsertIssues(si); err != nil {
		return err
	}
	ids := make([]int64, len(issues))
	for i, issue := range issues {
		ids[i] = issue.ID
	}
	if err := s.store.DeleteStaleIssues(ids); err != nil {
		return err
	}
	fmt.Println(styles.Success.Render(fmt.Sprintf("%d issues", len(issues))))
	return nil
}

func (s *Syncer) syncIssuesIncremental(ctx context.Context) error {
	fmt.Print("  syncing issues... ")
	lastSynced, err := s.store.GetLastSynced("issues")
	if err != nil {
		return err
	}
	var after *time.Time
	if !lastSynced.IsZero() {
		after = &lastSynced
	}
	issues, err := s.client.AllIssues(ctx, after)
	if err != nil {
		return err
	}
	if len(issues) > 0 {
		si := convertIssues(issues)
		if err := s.store.UpsertIssues(si); err != nil {
			return err
		}
	}
	fmt.Println(styles.Success.Render(fmt.Sprintf("%d issues", len(issues))))
	return nil
}

// SyncGroupIssues fetches issues for every stored group and links them.
func (s *Syncer) SyncGroupIssues(ctx context.Context) error {
	groups, err := s.store.ListGroups()
	if err != nil {
		return fmt.Errorf("list groups: %w", err)
	}
	if len(groups) == 0 {
		fmt.Println(styles.Label.Render("  no groups in store — run 'sync groups' first"))
		return nil
	}

	lastSynced, err := s.store.GetLastSynced("group_issues")
	if err != nil {
		return err
	}
	var after *time.Time
	if !lastSynced.IsZero() {
		after = &lastSynced
	}

	var totalIssues int
	for _, g := range groups {
		fmt.Printf("  syncing issues for %s... ", styles.Value.Render(g.Name))
		issues, err := s.client.AllGroupIssues(ctx, g.ID, after)
		if err != nil {
			fmt.Println(styles.Error.Render("failed"))
			return fmt.Errorf("group %d: %w", g.ID, err)
		}
		if len(issues) > 0 {
			si := convertIssues(issues)
			if err := s.store.UpsertIssues(si); err != nil {
				return err
			}
			ids := make([]int64, len(issues))
			for i, issue := range issues {
				ids[i] = issue.ID
			}
			if err := s.store.LinkGroupIssues(g.ID, ids); err != nil {
				return err
			}
		}
		fmt.Println(styles.Success.Render(fmt.Sprintf("%d issues", len(issues))))
		totalIssues += len(issues)
	}

	if err := s.store.SetLastSynced("group_issues", time.Now().UTC()); err != nil {
		return err
	}
	fmt.Println(styles.Success.Render(fmt.Sprintf("  total: %d issues across %d groups", totalIssues, len(groups))))
	return nil
}

// ShowStatus prints the last sync time for each resource type.
func (s *Syncer) ShowStatus() error {
	resources := []string{"user", "groups", "projects", "issues", "group_issues"}
	fmt.Println(styles.Title.Render("Sync Status"))
	for _, res := range resources {
		last, err := s.store.GetLastSynced(res)
		if err != nil {
			return err
		}
		ts := "never"
		if !last.IsZero() {
			ts = last.Local().Format("2006-01-02 15:04:05")
		}
		fmt.Printf("  %s %s\n",
			styles.Label.Render(fmt.Sprintf("%-10s", res)),
			styles.Value.Render(ts))
	}

	// Check if full sync is stale (>7 days).
	last, _ := s.store.GetLastFullSync("groups")
	if !last.IsZero() && time.Since(last) > 7*24*time.Hour {
		fmt.Println(styles.Error.Render("  hint: last full sync > 7 days ago — consider running 'sync full'"))
	}
	return nil
}

// NeedsFullSync returns true if the database appears to need a full sync
// (e.g. first run or stale data).
func (s *Syncer) NeedsFullSync() bool {
	empty, err := s.store.IsEmpty()
	if err != nil {
		return true
	}
	return empty
}
