// cmd/init.go
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type Config struct {
	NextID        int    `json:"next_id"`
	GitHubUser    string `json:"github_user,omitempty"`
	DefaultPublic bool   `json:"default_public"`
	BlogPath      string `json:"blog_path"`
	RepoName      string `json:"repo_name"`
}

type initModel struct {
	step        int
	blogName    textinput.Model
	blogPath    textinput.Model
	createRepo  bool
	currentUser string
	err         error
	quitting    bool
}

var initCmd = &cobra.Command{
	Use:   "init [blog-name]",
	Short: "Initialize a new gblog project",
	Long: `Initialize a new gblog project with automatic repository setup.

This creates a new blog repository, sets up the directory structure,
and configures everything needed to start your gist-powered blog.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return initializeBlogDirect(args[0])
		}
		return initializeBlogInteractive()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func initializeBlogInteractive() error {
	// Get current user for defaults
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	username := currentUser.Username
	if username == "" {
		username = "user"
	}

	m := initModel{
		step:        0,
		currentUser: username,
	}

	// Initialize blog name input
	m.blogName = textinput.New()
	m.blogName.Placeholder = fmt.Sprintf("gblog-%s", username)
	m.blogName.Focus()
	m.blogName.CharLimit = 100
	m.blogName.Width = 50

	// Initialize blog path input
	homeDir, _ := os.UserHomeDir()
	defaultPath := filepath.Join(homeDir, fmt.Sprintf("gblog-%s", username))

	m.blogPath = textinput.New()
	m.blogPath.Placeholder = defaultPath
	m.blogPath.CharLimit = 200
	m.blogPath.Width = 70

	m.createRepo = true // default

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	if finalModel.(initModel).quitting {
		fmt.Println("Cancelled.")
		return nil
	}

	return createBlogProject(finalModel.(initModel))
}

func initializeBlogDirect(blogName string) error {
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	homeDir, _ := os.UserHomeDir()
	blogPath := filepath.Join(homeDir, blogName)

	m := initModel{
		currentUser: currentUser.Username,
		createRepo:  true,
	}
	m.blogName = textinput.New()
	m.blogName.SetValue(blogName)
	m.blogPath = textinput.New()
	m.blogPath.SetValue(blogPath)

	return createBlogProject(m)
}

func (m initModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m initModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			switch m.step {
			case 0: // blog name step
				blogName := strings.TrimSpace(m.blogName.Value())
				if blogName == "" {
					blogName = m.blogName.Placeholder
					m.blogName.SetValue(blogName)
				}
				m.step = 1
				m.blogPath.Focus()
				m.blogName.Blur()
				m.err = nil
				return m, nil
			case 1: // blog path step
				blogPath := strings.TrimSpace(m.blogPath.Value())
				if blogPath == "" {
					blogPath = m.blogPath.Placeholder
					m.blogPath.SetValue(blogPath)
				}
				m.step = 2
				m.blogPath.Blur()
				return m, nil
			case 2: // create repo step
				return m, tea.Quit
			}
		case "y", "Y":
			if m.step == 2 {
				m.createRepo = true
				return m, tea.Quit
			}
		case "n", "N":
			if m.step == 2 {
				m.createRepo = false
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	switch m.step {
	case 0:
		m.blogName, cmd = m.blogName.Update(msg)
	case 1:
		m.blogPath, cmd = m.blogPath.Update(msg)
	}

	return m, cmd
}

func (m initModel) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("🚀 Initialize New Blog"))
	s.WriteString("\n")

	switch m.step {
	case 0:
		s.WriteString("What should your blog be called?\n\n")
		s.WriteString(inputStyle.Render(m.blogName.View()))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("Press Enter for default or type a custom name"))
	case 1:
		s.WriteString(fmt.Sprintf("Blog name: %s\n\n", m.blogName.Value()))
		s.WriteString("Where should your blog be created?\n\n")
		s.WriteString(inputStyle.Render(m.blogPath.View()))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("Press Enter for default location or specify custom path"))
	case 2:
		s.WriteString(fmt.Sprintf("Blog name: %s\n", m.blogName.Value()))
		s.WriteString(fmt.Sprintf("Location: %s\n", m.blogPath.Value()))
		s.WriteString("\nCreate GitHub repository? (y/n): ")
	}

	if m.err != nil {
		s.WriteString("\n\n")
		s.WriteString(errorStyle.Render(m.err.Error()))
	}

	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("Press Ctrl+C or Esc to cancel"))

	return s.String()
}

func createBlogProject(m initModel) error {
	blogName := m.blogName.Value()
	blogPath := m.blogPath.Value()

	fmt.Printf("🚀 Creating blog project: %s\n", blogName)
	fmt.Printf("📁 Location: %s\n", blogPath)

	// Create blog directory
	if err := os.MkdirAll(blogPath, 0755); err != nil {
		return fmt.Errorf("failed to create blog directory: %w", err)
	}

	// Change to blog directory
	if err := os.Chdir(blogPath); err != nil {
		return fmt.Errorf("failed to change to blog directory: %w", err)
	}

	// Initialize git repository
	fmt.Println("📋 Initializing git repository...")
	if err := runCommand("git", "init"); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Create blog structure
	if err := createBlogStructure(blogName); err != nil {
		return err
	}

	// Create initial commit
	fmt.Println("💾 Creating initial commit...")
	if err := runCommand("git", "add", "."); err != nil {
		return fmt.Errorf("failed to add files to git: %w", err)
	}

	if err := runCommand("git", "commit", "-m", "Initial commit: Initialize gblog"); err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	// Create GitHub repository if requested
	if m.createRepo {
		fmt.Println("🌐 Creating GitHub repository...")
		if err := createGitHubRepo(blogName); err != nil {
			fmt.Printf("⚠️  Could not create GitHub repository: %v\n", err)
			fmt.Println("You can create it manually later with: gh repo create")
		} else {
			fmt.Println("📤 Pushing to GitHub...")
			if err := runCommand("git", "push", "-u", "origin", "main"); err != nil {
				fmt.Printf("⚠️  Could not push to GitHub: %v\n", err)
			}
		}
	}

	fmt.Printf("✅ Blog '%s' created successfully!\n", blogName)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. cd %s\n", blogPath)
	fmt.Println("  2. gblog new              # Create your first post")
	fmt.Println("  3. gblog publish 0001     # Publish when ready")
	fmt.Println()
	fmt.Printf("📂 Blog directory: %s\n", blogPath)

	return nil
}

func createBlogStructure(blogName string) error {
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
		BlogPath:      ".",
		RepoName:      blogName,
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

	// Create README
	readmeContent := fmt.Sprintf(`# %s

A gist-powered blog created with [gblog](https://github.com/onprema/gblog).

## Posts

This repository contains my blog posts, each published as a GitHub Gist.

Posts are organized with descriptive filenames (e.g., `+"`getting-started-with-go.md`"+`) rather than generic names.

## Usage

- Create new post: `+"`gblog new`"+`
- List posts: `+"`gblog list`"+`
- Edit post: `+"`gblog edit <id>`"+`
- Publish post: `+"`gblog publish <id>`"+`
- Update existing gist: `+"`gblog publish <id> --update`"+`
- Export all: `+"`gblog export`"+`

## Posts Directory

All posts are organized in the `+"`posts/`"+` directory with the format `+"`XXXX-post-title/`"+`.
Each post contains a descriptively named markdown file and any auxiliary files.

## Workflow

1. `+"`gblog new`"+` - Create post with interactive prompts
2. `+"`gblog edit <id>`"+` - Open directory to write content
3. `+"`git add . && git commit`"+` - Version control your changes
4. `+"`gblog publish <id>`"+` - Publish to GitHub Gists
5. `+"`gblog publish <id> --update`"+` - Update gist after changes
`, blogName)

	if err := os.WriteFile("README.md", []byte(readmeContent), 0644); err != nil {
		return fmt.Errorf("failed to create README: %w", err)
	}

	// Create .gitignore for blog repo
	blogGitignore := `# gblog private posts will be added here automatically

# OS generated files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# Export files
*.zip
`

	if err := os.WriteFile(".gitignore", []byte(blogGitignore), 0644); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}

	return nil
}

func createGitHubRepo(repoName string) error {
	// Check if gh CLI is available and authenticated
	if err := runCommand("gh", "auth", "status"); err != nil {
		return fmt.Errorf("GitHub CLI not authenticated. Run 'gh auth login' first")
	}

	// Create the repository
	description := fmt.Sprintf("A gist-powered blog created with gblog")
	return runCommand("gh", "repo", "create", repoName, "--public", "--description", description, "--source=.", "--remote=origin", "--push")
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func initializeBlog() error {
	// Legacy function - redirect to interactive
	return initializeBlogInteractive()
}
