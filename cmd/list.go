// cmd/list.go
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	tableHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7C3AED")).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color("#7C3AED"))

	tableRowStyle = lipgloss.NewStyle().
		PaddingRight(2)

	publishedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#22C55E")).
		Bold(true)

	draftStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F59E0B")).
		Bold(true)

	privateStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EF4444")).
		Bold(true)
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

	// Display table
	fmt.Println(titleStyle.Render("📝 Blog Posts"))
	fmt.Println()

	// Headers
	headers := []string{"ID", "Title", "Status", "Visibility", "Created", "Gist URL"}
	headerRow := ""
	for i, header := range headers {
		width := getColumnWidth(i)
		headerRow += tableHeaderStyle.Width(width).Render(header)
		if i < len(headers)-1 {
			headerRow += " "
		}
	}
	fmt.Println(headerRow)

	// Rows
	for _, post := range posts {
		row := ""

		// ID
		row += tableRowStyle.Width(getColumnWidth(0)).Render(post.Meta.ID)
		row += " "

		// Title (truncated if too long)
		title := post.Meta.Title
		if len(title) > 30 {
			title = title[:27] + "..."
		}
		row += tableRowStyle.Width(getColumnWidth(1)).Render(title)
		row += " "

		// Status
		var status string
		if post.Meta.GistID != "" {
			status = publishedStyle.Render("Published")
		} else {
			status = draftStyle.Render("Draft")
		}
		row += tableRowStyle.Width(getColumnWidth(2)).Render(status)
		row += " "

		// Visibility
		var visibility string
		if post.Meta.Public {
			visibility = "Public"
		} else {
			visibility = privateStyle.Render("Private")
		}
		row += tableRowStyle.Width(getColumnWidth(3)).Render(visibility)
		row += " "

		// Created date
		created := post.Meta.CreatedAt.Format("2006-01-02")
		row += tableRowStyle.Width(getColumnWidth(4)).Render(created)
		row += " "

		// Gist URL
		gistURL := post.Meta.GistURL
		if gistURL == "" {
			gistURL = "-"
		} else if len(gistURL) > 40 {
			gistURL = gistURL[:37] + "..."
		}
		row += tableRowStyle.Width(getColumnWidth(5)).Render(gistURL)

		fmt.Println(row)
	}

	fmt.Println()
	fmt.Printf("Total posts: %d\n", len(posts))

	// Count stats
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

	fmt.Printf("Published: %d, Drafts: %d, Private: %d\n", published, len(posts)-published, private)

	return nil
}

func getColumnWidth(col int) int {
	widths := []int{4, 32, 10, 10, 12, 42}
	if col < len(widths) {
		return widths[col]
	}
	return 20
