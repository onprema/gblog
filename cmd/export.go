package cmd

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export [output-file]",
	Short: "Export all posts to a zip file",
	Long: `Export all blog posts (public and private) to a zip file.

The exported archive will contain all posts organized by date,
including all markdown files and auxiliary files.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		outputFile := "gblog-export.zip"
		if len(args) > 0 {
			outputFile = args[0]
		}
		return exportPosts(outputFile)
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
}

func exportPosts(outputFile string) error {
	// Check if gblog is initialized
	if _, err := os.Stat(".gblog/config.json"); os.IsNotExist(err) {
		return fmt.Errorf("gblog not initialized. Run 'gblog init' first")
	}

	// Read posts directory
	postsDir := "posts"
	if _, err := os.Stat(postsDir); os.IsNotExist(err) {
		return fmt.Errorf("no posts directory found")
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
		return fmt.Errorf("no posts found to export")
	}

	// Sort posts by creation date
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Meta.CreatedAt.Before(posts[j].Meta.CreatedAt)
	})

	// Create zip file
	zipFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	fmt.Printf("📦 Exporting %d posts to %s...\n", len(posts), outputFile)

	// Add each post to the zip
	for _, post := range posts {
		postPath := filepath.Join(postsDir, post.Dir)

		// Create directory structure based on creation date
		createdDate := post.Meta.CreatedAt.Format("2006/01/02")
		zipDirPath := filepath.Join("posts", createdDate, post.Dir)

		fmt.Printf("  📁 Adding %s (%s)...\n", post.Meta.Title, post.Meta.ID)

		// Add all files in the post directory
		err := filepath.Walk(postPath, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			// Calculate relative path within the post directory
			relPath, err := filepath.Rel(postPath, filePath)
			if err != nil {
				return err
			}

			// Create the file path in the zip
			zipFilePath := filepath.Join(zipDirPath, relPath)
			zipFilePath = filepath.ToSlash(zipFilePath) // Ensure forward slashes in zip

			// Create the file in the zip
			zipFileWriter, err := zipWriter.Create(zipFilePath)
			if err != nil {
				return fmt.Errorf("failed to create file in zip: %w", err)
			}

			// Copy file contents
			fileReader, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer fileReader.Close()

			_, err = io.Copy(zipFileWriter, fileReader)
			if err != nil {
				return fmt.Errorf("failed to copy file contents: %w", err)
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to add post %s to zip: %w", post.Meta.ID, err)
		}
	}

	// Add export metadata
	exportMeta := struct {
		ExportedAt time.Time `json:"exported_at"`
		TotalPosts int       `json:"total_posts"`
		Posts      []struct {
			ID        string    `json:"id"`
			Title     string    `json:"title"`
			Public    bool      `json:"public"`
			CreatedAt time.Time `json:"created_at"`
			GistURL   string    `json:"gist_url,omitempty"`
		} `json:"posts"`
	}{
		ExportedAt: time.Now(),
		TotalPosts: len(posts),
	}

	for _, post := range posts {
		exportMeta.Posts = append(exportMeta.Posts, struct {
			ID        string    `json:"id"`
			Title     string    `json:"title"`
			Public    bool      `json:"public"`
			CreatedAt time.Time `json:"created_at"`
			GistURL   string    `json:"gist_url,omitempty"`
		}{
			ID:        post.Meta.ID,
			Title:     post.Meta.Title,
			Public:    post.Meta.Public,
			CreatedAt: post.Meta.CreatedAt,
			GistURL:   post.Meta.GistURL,
		})
	}

	// Add export metadata file
	metaWriter, err := zipWriter.Create("export-metadata.json")
	if err != nil {
		return fmt.Errorf("failed to create metadata file in zip: %w", err)
	}

	encoder := json.NewEncoder(metaWriter)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(exportMeta); err != nil {
		return fmt.Errorf("failed to write export metadata: %w", err)
	}

	fmt.Printf("✅ Export completed successfully!\n")
	fmt.Printf("📦 Archive: %s\n", outputFile)
	fmt.Printf("📊 Total posts: %d\n", len(posts))

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

	fmt.Printf("📈 Published: %d, Drafts: %d, Private: %d\n", published, len(posts)-published, private)

	return nil
}
