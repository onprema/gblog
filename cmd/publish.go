// cmd/publish.go
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var publishCmd = &cobra.Command{
	Use:   "publish <post-id>",
	Short: "Publish a post to GitHub Gists",
	Long: `Publish a blog post to GitHub Gists.

This command will upload all files in the post directory to a new gist
and open it in your default browser. Use --update to update an existing gist.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		update, _ := cmd.Flags().GetBool("update")
		return publishPost(args[0], update)
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)
	publishCmd.Flags().BoolP("update", "u", false, "Update existing gist instead of creating new one")
}

func publishPost(postID string, update bool) error {
	// Find post directory
	postDir, err := findPostDir(postID)
	if err != nil {
		return err
	}

	// Load metadata
	metaPath := filepath.Join(postDir, ".meta.json")
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		return fmt.Errorf("failed to read post metadata: %w", err)
	}

	var meta PostMeta
	if err := json.Unmarshal(metaData, &meta); err != nil {
		return fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Check if already published and handle accordingly
	if meta.GistID != "" && !update {
		fmt.Printf("⚠️  Post already published: %s\n", meta.GistURL)
		fmt.Println("Use 'gblog publish --update' to update the existing gist.")
		return nil
	}

	// Check gh CLI authentication
	if err := checkGHAuth(); err != nil {
		return err
	}

	var gistURL, gistID string

	if meta.GistID != "" && update {
		// Update existing gist
		gistURL, gistID, err = updateExistingGist(postDir, &meta)
		if err != nil {
			return err
		}
		fmt.Printf("✅ Updated existing gist!\n")
	} else {
		// Create new gist
		gistURL, gistID, err = createNewGist(postDir, &meta)
		if err != nil {
			return err
		}
		fmt.Printf("✅ Published successfully!\n")
	}

	// Update metadata with gist info
	meta.GistID = gistID
	meta.GistURL = gistURL

	metaFile, err := os.Create(metaPath)
	if err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}
	defer metaFile.Close()

	encoder := json.NewEncoder(metaFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(meta); err != nil {
		return fmt.Errorf("failed to write updated metadata: %w", err)
	}

	fmt.Printf("🔗 Gist URL: %s\n", gistURL)
	fmt.Printf("📝 Gist ID: %s\n", gistID)

	// Open in browser
	fmt.Println("🌐 Opening in browser...")
	if err := openInBrowser(gistURL); err != nil {
		fmt.Printf("⚠️  Could not open browser automatically: %v\n", err)
		fmt.Printf("Please visit: %s\n", gistURL)
	}

	return nil
}

func createNewGist(postDir string, meta *PostMeta) (string, string, error) {
	// Prepare gist creation command
	args := []string{"gist", "create"}

	if meta.Public {
		args = append(args, "--public")
	}

	if meta.Description != "" {
		args = append(args, "--desc", meta.Description)
	}

	// Add filename arguments for all files in the directory
	gistFiles, err := getGistFiles(postDir)
	if err != nil {
		return "", "", err
	}

	if len(gistFiles) == 0 {
		return "", "", fmt.Errorf("no files found to publish in %s", postDir)
	}

	args = append(args, gistFiles...)

	fmt.Printf("📤 Publishing post '%s'...\n", meta.Title)
	fmt.Printf("Files: %v\n", gistFiles)

	// Execute gh gist create
	cmd := exec.Command("gh", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return "", "", fmt.Errorf("failed to create gist: %s", string(exitError.Stderr))
		}
		return "", "", fmt.Errorf("failed to create gist: %w", err)
	}

	gistURL := strings.TrimSpace(string(output))

	// Extract gist ID from URL
	parts := strings.Split(gistURL, "/")
	if len(parts) == 0 {
		return "", "", fmt.Errorf("invalid gist URL returned: %s", gistURL)
	}
	gistID := parts[len(parts)-1]

	return gistURL, gistID, nil
}

func updateExistingGist(postDir string, meta *PostMeta) (string, string, error) {
	// Get all files to update
	gistFiles, err := getGistFiles(postDir)
	if err != nil {
		return "", "", err
	}

	if len(gistFiles) == 0 {
		return "", "", fmt.Errorf("no files found to update in %s", postDir)
	}

	fmt.Printf("📤 Updating existing gist '%s'...\n", meta.Title)
	fmt.Printf("Files: %v\n", gistFiles)

	// Prepare update command
	args := []string{"gist", "edit", meta.GistID}
	args = append(args, gistFiles...)

	// Execute gh gist edit
	cmd := exec.Command("gh", args...)
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return "", "", fmt.Errorf("failed to update gist: %s", string(exitError.Stderr))
		}
		return "", "", fmt.Errorf("failed to update gist: %w", err)
	}

	// Return existing URL and ID
	return meta.GistURL, meta.GistID, nil
}

func getGistFiles(postDir string) ([]string, error) {
	files, err := os.ReadDir(postDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read post directory: %w", err)
	}

	var gistFiles []string
	for _, file := range files {
		if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
			continue // Skip directories and hidden files like .meta.json
		}

		filePath := filepath.Join(postDir, file.Name())
		gistFiles = append(gistFiles, filePath)
	}

	return gistFiles, nil
}

func findPostDir(postID string) (string, error) {
	postsDir := "posts"
	entries, err := os.ReadDir(postsDir)
	if err != nil {
		return "", fmt.Errorf("failed to read posts directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), postID+"-") {
			return filepath.Join(postsDir, entry.Name()), nil
		}
	}

	return "", fmt.Errorf("post with ID %s not found", postID)
}

func checkGHAuth() error {
	cmd := exec.Command("gh", "auth", "status")
	if err := cmd.Run(); err != nil {
		fmt.Println("🔐 GitHub CLI authentication required.")
		fmt.Println("Please run: gh auth login")
		return fmt.Errorf("GitHub CLI not authenticated")
	}
	return nil
}

func openInBrowser(url string) error {
	var cmd *exec.Cmd

	switch {
	case isCommandAvailable("open"): // macOS
		cmd = exec.Command("open", url)
	case isCommandAvailable("xdg-open"): // Linux
		cmd = exec.Command("xdg-open", url)
	case isCommandAvailable("cmd"): // Windows
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("no browser opening command available")
	}

	return cmd.Run()
}

func isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}
