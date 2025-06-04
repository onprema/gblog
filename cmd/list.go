// cmd/list.go
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	listTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			Margin(1, 0)

	publishedColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#22C55E"))
	draftColor     = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B"))
	privateColor   = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
)

type PostInfo struct {
	Meta PostMeta
	Dir  string
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all blog posts",
	Long: `List all blog posts with their status and information.

Shows post ID, title, status (draft/published), visibility (public/private),
and creation date.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listPosts()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func listPosts() error {
	// Check if gblog is initialized
	if _, err := os.Stat(".gblog/config.json"); os.IsNotExist(err) {
		return fmt.Errorf("gblog not initialized. Run 'gblog init' first")
	}

	// Read posts directory
	postsDir := "posts"
	if _, err := os.Stat(postsDir); os.IsNotExist(err) {
		fmt.Println("No posts found. Create your first post with 'gblog new'")
		return nil
	}

	entries, err := os.ReadDir(postsDir)
	if err != nil {
		return fmt.Errorf("failed to read posts directory: %w", err)
	}

	var posts []PostInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		metaPath := filepath.Join(postsDir, entry.Name(), ".meta.json")
		metaData, err := os.ReadFile(metaPath)
		if err != nil {
			fmt.Printf("Warning: could not read metadata for %s: %v\n", entry.Name(), err)
			continue
		}

		var meta PostMeta
		if err := json.Unmarshal(metaData, &meta); err != nil {
			fmt.Printf("Warning: could not parse metadata for %s: %v\n", entry.Name(), err)
			continue
		}

		posts = append(posts, PostInfo{
			Meta: meta,
			Dir:  entry.Name(),
		})
	}

	if len(posts) == 0 {
		fmt.Println("No posts found. Create your first post with 'gblog new'")
		return nil
	}

	// Sort posts by ID (newest first)
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Meta.ID > posts[j].Meta.ID
	})

	// Display header
	fmt.Println(listTitleStyle.Render("📝 Blog Posts"))
	fmt.Println()

	// Simple table without complex formatting
	fmt.Printf("%-4s %-35s %-12s %-10s %-12s %s\n",
		"ID", "Title", "Status", "Visibility", "Created", "Gist URL")
	fmt.Println(strings.Repeat("-", 120))

	// Table rows
	for _, post := range posts {
		// Truncate title if too long
		title := post.Meta.Title
		if len(title) > 33 {
			title = title[:30] + "..."
		}

		// Status
		status := "Draft"
		statusColor := draftColor
		if post.Meta.GistID != "" {
			status = "Published"
			statusColor = publishedColor
		}

		// Visibility
		visibility := "Public"
		visibilityColor := lipgloss.NewStyle() // no color for public
		if !post.Meta.Public {
			visibility = "Private"
			visibilityColor = privateColor
		}

		// Created date
		created := post.Meta.CreatedAt.Format("2006-01-02")

		// Gist URL
		gistURL := "-"
		if post.Meta.GistURL != "" {
			gistURL = post.Meta.GistURL
			if len(gistURL) > 45 {
				gistURL = gistURL[:42] + "..."
			}
		}

		// Print row with colors
		fmt.Printf("%-4s %-35s %-12s %-10s %-12s %s\n",
			post.Meta.ID,
			title,
			statusColor.Render(status),
			visibilityColor.Render(visibility),
			created,
			gistURL)
	}

	fmt.Println()

	// Stats
	published := 0
	private := 0
	for _, post := range posts {
		if post.Meta.GistID != "" {
			published++
		}
		if !post.Meta.Public {
			private++
		}
	}

	fmt.Printf("Total: %d | Published: %d | Drafts: %d | Private: %d\n",
		len(posts), published, len(posts)-published, private)

	return nil
}
