// cmd/edit.go
package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <post-id>",
	Short: "Open a post directory for editing",
	Long: `Open a post directory in your default file manager or editor.

This will open the post directory so you can edit the markdown file
and add any auxiliary files before publishing.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return editPost(args[0])
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func editPost(postID string) error {
	// Find post directory
	postDir, err := findPostDir(postID)
	if err != nil {
		return err
	}

	fmt.Printf("📁 Opening post directory: %s\n", postDir)

	// Try to open the directory in the file manager
	if err := openDirectory(postDir); err != nil {
		fmt.Printf("⚠️  Could not open file manager: %v\n", err)
		fmt.Printf("📂 Post directory: %s\n", postDir)
		fmt.Printf("💡 You can manually navigate to this directory to edit your files\n")
		return nil
	}

	fmt.Printf("✅ Opened in file manager\n")
	fmt.Printf("💡 Edit your files and run 'gblog publish %s' when ready\n", postID)

	return nil
}

func openDirectory(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("open", path)
	case "linux":
		cmd = exec.Command("xdg-open", path)
	case "windows":
		cmd = exec.Command("explorer", path)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return cmd.Run()
}
