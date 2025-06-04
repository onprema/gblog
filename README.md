# gblog

A gist-powered blog CLI tool that turns GitHub Gists into your personal blog platform.

## Overview

gblog lets you write blog posts in markdown, add auxiliary files (code examples, images, etc.), and publish them as GitHub Gists. Your blog becomes a collection of organized, searchable, and shareable code snippets and thoughts.

**Two-Repository Design:**
- **gblog** - The CLI tool (this repo)
- **your-blog-name** - Your blog content repository (created automatically)

## Features

- **Automated blog setup** with GitHub repository creation
- **Interactive post creation** with beautiful CLI interface
- **Descriptive filenames** - posts create meaningful .md files, not generic "post.md"
- **Organized post management** with auto-generated IDs
- **GitHub Gists integration** for publishing and sharing
- **Gist updates** - easily update existing gists with changes
- **Public/private post support** with automatic .gitignore management
- **Full version control** for your blog posts
- **Export functionality** to backup all your posts
- **Cross-platform** support (macOS, Linux, Windows)

## Prerequisites

- Go 1.21 or later
- [GitHub CLI](https://cli.github.com/) installed and authenticated
- Git (for version control)

## Installation

### From Source

```bash
git clone https://github.com/onprema/gblog
cd gblog
make install
```

### Direct Install

```bash
go install github.com/onprema/gblog@latest
```

## Quick Start

### 1. Create Your Blog

```bash
# Interactive setup (recommended)
gblog init

# Or specify blog name directly
gblog init my-tech-blog
```

This will:
- Create a new directory (default: `~/gblog-username`)
- Initialize a git repository
- Create GitHub repository automatically
- Set up blog structure
- Make initial commit and push

### 2. Create Your First Post

```bash
cd ~/gblog-username  # or your custom path
gblog new
```

Interactive prompts for:
- Post title
- Description (optional)
- Public/private visibility

### 3. Write Your Content

```bash
gblog edit 0001
```

This opens your post directory. Edit `post.md` and add any auxiliary files.

### 4. Publish to Gists

```bash
gblog publish 0001
```

Creates a gist and opens it in your browser automatically.

## Commands

| Command | Description |
|---------|-------------|
| `gblog init [name]` | Create new blog with repository setup |
| `gblog new` | Create a new blog post interactively |
| `gblog list` | List all blog posts with status |
| `gblog edit <id>` | Open post directory for editing |
| `gblog publish <id>` | Publish post to GitHub Gists |
| `gblog publish <id> --update` | Update existing gist with changes |
| `gblog export [file]` | Export all posts to zip file |

## Project Structure

**Tool Repository (gblog):**
```
gblog/
├── cmd/              # CLI commands
├── main.go          # Entry point
├── go.mod           # Dependencies
├── Makefile         # Build automation
└── README.md        # This file
```

**Blog Repository (created by init):**
```
my-tech-blog/
├── .gblog/
│   └── config.json                    # Blog configuration
├── posts/
│   ├── 0001-my-first-post/
│   │   ├── .meta.json                 # Post metadata
│   │   ├── my-first-post.md           # Main content (descriptive filename)
│   │   └── example.go                 # Auxiliary files
│   └── 0002-iam-gcp-vs-aws/
│       ├── .meta.json
│       ├── iam-gcp-vs-aws.md          # Descriptive filename based on title
│       └── comparison-table.json
├── .gitignore                         # Auto-updated for private posts
└── README.md                          # Blog description
```

## Example Workflow

```bash
# Install gblog
go install github.com/onprema/gblog@latest

# Create your blog (interactive)
gblog init
# Enter blog name: my-tech-blog
# Enter location: /Users/john/my-tech-blog (default)
# Create GitHub repo: y

# Navigate to your blog
cd ~/my-tech-blog

# Create first post
gblog new
# Title: "Getting Started with Go Generics"
# Description: "A practical guide to using generics in Go"
# Public: y

# Edit the post
gblog edit 0001
# Add content to getting-started-with-go-generics.md, include example files

# Commit your work
git add .
git commit -m "Add: Getting Started with Go Generics"
git push

# Publish to gist when ready
gblog publish 0001
# Opens: https://gist.github.com/yourusername/...

# Make updates to your post
gblog edit 0001
# Edit files, rename, add new files, etc.

# Update the existing gist
gblog publish 0001 --update
# Updates the same gist with your changes

# List all posts
gblog list

# Export backup
gblog export my-blog-backup.zip
```

## Blog Repository Features

- **Version controlled** - Full git history of all posts
- **GitHub hosted** - Automatic repository creation and push
- **Private post support** - Automatically added to .gitignore
- **Collaborative** - Others can contribute via pull requests
- **Portable** - Clone your blog anywhere
- **Multiple blogs** - Create separate repositories for different topics

## Post Metadata

Each post includes metadata in `.meta.json`:

```json
{
  "id": "0001",
  "title": "Getting Started with Go Generics",
  "description": "A practical guide to using generics in Go",
  "public": true,
  "created_at": "2025-06-04T10:30:00Z",
  "gist_id": "abc123...",
  "gist_url": "https://gist.github.com/yourusername/abc123..."
}
```

## Development

```bash
# Clone the tool repository
git clone https://github.com/onprema/gblog
cd gblog

# Install dependencies
go mod tidy

# Build
make build

# Run during development
make run ARGS="init test-blog"

# Install locally
make install
```

## Multiple Blogs

You can create multiple blogs for different purposes:

```bash
gblog init work-blog          # Professional content
gblog init personal-thoughts  # Personal posts
gblog init tutorials         # Educational content
```

Each gets its own repository and configuration.

## Privacy & Security

- **Private posts** are automatically added to `.gitignore` in your blog repo
- **Public posts** are committed to your blog repository and published as public gists
- **GitHub authentication** is handled by GitHub CLI
- **Full control** over your content with git version history
- **Backup redundancy** - content exists in your repo AND as gists

## Why Gist-Powered Blogging?

- **Version control** - Every post is versioned through Git and GitHub
- **Searchable** - Gists are indexed and searchable on GitHub (e.g., `user:onprema`)
- **Embeddable** - Easy to embed gists in documentation or other sites
- **Collaborative** - Others can fork, comment, and suggest improvements
- **Portable** - Your content lives in your own repositories
- **Developer-friendly** - Write in markdown with syntax highlighting
- **Zero hosting costs** - Uses GitHub's infrastructure

## Repository Separation

- **Tool development** happens in the `gblog` repository
- **Blog content** lives in separate repositories created by `gblog init`
- **Clean separation** ensures the tool remains reusable
- **Personal content** stays in your own repositories
