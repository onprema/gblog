// cmd/new.go
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

type PostMeta struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Public      bool      `json:"public"`
	CreatedAt   time.Time `json:"created_at"`
	GistID      string    `json:"gist_id,omitempty"`
	GistURL     string    `json:"gist_url,omitempty"`
}

type newPostModel struct {
	step        int
	title       textinput.Model
	description textinput.Model
	isPublic    bool
	err         error
	quitting    bool
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			Margin(1, 0)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Margin(1, 0)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new blog post",
	Long: `Create a new blog post with an interactive CLI.

This will prompt you for the post title, description, and visibility,
then create a new directory with the post files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runNewPost()
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}

func runNewPost() error {
	// Check if gblog is initialized
	if _, err := os.Stat(".gblog/config.json"); os.IsNotExist(err) {
		return fmt.Errorf("gblog not initialized. Run 'gblog init' first")
	}

	m := newPostModel{
		step: 0,
	}

	// Initialize title input
	m.title = textinput.New()
	m.title.Placeholder = "Enter your post title..."
	m.title.Focus()
	m.title.CharLimit = 100
	m.title.Width = 50

	// Initialize description input
	m.description = textinput.New()
	m.description.Placeholder = "Enter post description (optional)..."
	m.description.CharLimit = 200
	m.description.Width = 50

	m.isPublic = true // default

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	if finalModel.(newPostModel).quitting {
		fmt.Println("Cancelled.")
		return nil
	}

	return createPost(finalModel.(newPostModel))
}

func (m newPostModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m newPostModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			switch m.step {
			case 0: // title step
				if strings.TrimSpace(m.title.Value()) == "" {
					m.err = fmt.Errorf("title cannot be empty")
					return m, nil
				}
				m.step = 1
				m.description.Focus()
				m.title.Blur()
				m.err = nil
				return m, nil
			case 1: // description step
				m.step = 2
				m.description.Blur()
				return m, nil
			case 2: // public/private step
				return m, tea.Quit
			}
		case "y", "Y":
			if m.step == 2 {
				m.isPublic = true
				return m, tea.Quit
			}
		case "n", "N":
			if m.step == 2 {
				m.isPublic = false
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	switch m.step {
	case 0:
		m.title, cmd = m.title.Update(msg)
	case 1:
		m.description, cmd = m.description.Update(msg)
	}

	return m, cmd
}

func (m newPostModel) View() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("📝 Create New Blog Post"))
	s.WriteString("\n")

	switch m.step {
	case 0:
		s.WriteString("What's the title of your post?\n\n")
		s.WriteString(inputStyle.Render(m.title.View()))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("Press Enter to continue"))
	case 1:
		s.WriteString(fmt.Sprintf("Title: %s\n\n", m.title.Value()))
		s.WriteString("Post description (optional):\n\n")
		s.WriteString(inputStyle.Render(m.description.View()))
		s.WriteString("\n\n")
		s.WriteString(helpStyle.Render("Press Enter to continue (or leave empty)"))
	case 2:
		s.WriteString(fmt.Sprintf("Title: %s\n", m.title.Value()))
		if m.description.Value() != "" {
			s.WriteString(fmt.Sprintf("Description: %s\n", m.description.Value()))
		}
		s.WriteString("\nShould this post be public? (y/n): ")
	}

	if m.err != nil {
		s.WriteString("\n\n")
		s.WriteString(errorStyle.Render(m.err.Error()))
	}

	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("Press Ctrl+C or Esc to cancel"))

	return s.String()
}

func createPost(m newPostModel) error {
	// Load config
	configData, err := os.ReadFile(".gblog/config.json")
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Generate post ID and directory name
	postID := fmt.Sprintf("%04d", config.NextID)
	slug := slugify(m.title.Value())
	dirName := fmt.Sprintf("%s-%s", postID, slug)
	postDir := filepath.Join("posts", dirName)

	// Create post directory
	if err := os.MkdirAll(postDir, 0755); err != nil {
		return fmt.Errorf("failed to create post directory: %w", err)
	}

	// Create metadata file
	meta := PostMeta{
		ID:          postID,
		Title:       m.title.Value(),
		Description: m.description.Value(),
		Public:      m.isPublic,
		CreatedAt:   time.Now(),
	}

	metaPath := filepath.Join(postDir, ".meta.json")
	metaFile, err := os.Create(metaPath)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer metaFile.Close()

	encoder := json.NewEncoder(metaFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(meta); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	// Create markdown file with descriptive name
	mdFilename := fmt.Sprintf("%s.md", slug)
	mdPath := filepath.Join(postDir, mdFilename)
	mdContent := fmt.Sprintf("# %s\n\n", m.title.Value())
	if m.description.Value() != "" {
		mdContent += fmt.Sprintf("*%s*\n\n", m.description.Value())
	}
	mdContent += "Write your post content here...\n"

	if err := os.WriteFile(mdPath, []byte(mdContent), 0644); err != nil {
		return fmt.Errorf("failed to create markdown file: %w", err)
	}

	// Update config with next ID
	config.NextID++
	configFile, err := os.Create(".gblog/config.json")
	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}
	defer configFile.Close()

	encoder = json.NewEncoder(configFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to write updated config: %w", err)
	}

	// Add to .gitignore if private
	if !m.isPublic {
		gitignoreEntry := fmt.Sprintf("posts/%s/\n", dirName)
		file, err := os.OpenFile(".gitignore", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Warning: could not update .gitignore: %v\n", err)
		} else {
			defer file.Close()
			file.WriteString(gitignoreEntry)
		}
	}

	fmt.Printf("✅ Created new post: %s\n", dirName)
	fmt.Printf("📁 Directory: posts/%s/\n", dirName)
	fmt.Printf("📝 Edit your post: posts/%s/%s.md\n", dirName, slug)
	if !m.isPublic {
		fmt.Printf("🔒 This post is private and added to .gitignore\n")
	}
	fmt.Printf("\nWhen ready, publish with: gblog publish %s\n", postID)

	return nil
}

func slugify(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces and special chars with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")

	// Remove leading/trailing hyphens
	s = strings.Trim(s, "-")

	// Limit length
	if len(s) > 50 {
		s = s[:50]
	}

	return s
}
