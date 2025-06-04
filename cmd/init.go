// cmd/init.go
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

type Config struct {
	NextID        int    `json:"next_id"`
	GitHubUser    string `json:"github_user,omitempty"`
	DefaultPublic bool   `json:"default_public"`
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new gblog project",
	Long: `Initialize a new gblog project in the current directory.

This creates the necessary directory structure and configuration files
to start your gist-powered blog.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return initializeBlog()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func initializeBlog() error {
	// Check if already initialized
	if _, err := os.Stat(".gblog"); err == nil {
		return fmt.Errorf("gblog already initialized in this directory")
	}

	// Create .gblog directory
	if err := os.MkdirAll(".gblog", 0755); err != nil {
		return fmt.Errorf("failed to create .gblog directory: %w", err)
	}

	// Create posts directory
	if err := os.MkdirAll("posts", 0755); err != nil {
		return fmt.Errorf("failed to create posts directory: %w", err)
	}

	// Create initial config
	config := Config{
		NextID:        1,
		DefaultPublic: true,
	}

	configPath := filepath.Join(".gblog", "config.json")
	configFile, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer configFile.Close()

	encoder := json.NewEncoder(configFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Create .gitignore entries for private posts
	gitignorePath := ".gitignore"
	gitignoreContent := "\n# gblog private posts\n"

	// Check if .gitignore exists
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		// Create new .gitignore
		if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
			fmt.Printf("Warning: could not create .gitignore: %v\n", err)
		}
	} else {
		// Append to existing .gitignore
		file, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Warning: could not update .gitignore: %v\n", err)
		} else {
			defer file.Close()
			file.WriteString(gitignoreContent)
		}
	}

	fmt.Println("✅ gblog initialized successfully!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Run 'gblog new' to create your first post")
	fmt.Println("  2. Write your content in the generated directory")
	fmt.Println("  3. Run 'gblog publish <id>' to publish to GitHub Gists")
	fmt.Println()
	fmt.Println("Directory structure created:")
	fmt.Println("  .gblog/config.json  - Configuration file")
	fmt.Println("  posts/              - Your blog posts will go here")

	return nil
}
