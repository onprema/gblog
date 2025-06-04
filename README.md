# gblog

A gist-powered blog CLI tool that turns GitHub Gists into your personal blog platform.

## Overview

gblog lets you write blog posts in markdown, add auxiliary files (code examples, images, etc.), and publish them as GitHub Gists. Your blog becomes a collection of organized, searchable, and shareable code snippets and thoughts.

## Features

- **Interactive post creation** with beautiful CLI interface
- **Organized post management** with auto-generated IDs
- **GitHub Gists integration** for publishing and sharing
- **Public/private post support** with automatic .gitignore management
- **Export functionality** to backup all your posts
- **Cross-platform** support (macOS, Linux, Windows)

## Prerequisites

- Go 1.21 or later
- [GitHub CLI](https://cli.github.com/) installed and authenticated
- Git (for version control)

## Installation

### From Source

```bash
git clone https://github.com/your-username/gblog
cd gblog
make install
```

### Direct Install

```bash
go install github.com/your-username/gblog@latest
```

## Quick Start

1. **Initialize your blog**
   ```bash
   gblog init
   ```

2. **Create your first post**
   ```bash
   gblog new
   ```
   This opens an interactive prompt for title, description, and visibility.

3. **Edit your post**
   ```bash
   gblog edit 0001
   ```
   Add content to `posts/0001-your-post-title/post.md` and any auxiliary files.

4. **Publish to GitHub Gists**
   ```bash
   gblog publish 0001
   ```
   This creates a gist and opens it in your browser.

## Commands

| Command | Description |
|---------|-------------|
| `gblog init` | Initialize a new gblog project |
| `gblog new` | Create a new blog post interactively |
| `gblog list` | List all blog posts with status |
| `gblog edit <id>` | Open post directory for editing |
| `gblog publish <id>` | Publish post to GitHub Gists |
| `gblog export [file]` | Export all posts to zip file |

## Project Structure

```
my-blog/
├── .gblog/
│   └── config.json          # gblog configuration
├── posts/
│   ├── 0001-my-first-post/
│   │   ├── .meta.json       # Post metadata
│   │   ├── post.md          # Main content
│   │   └── example.go       # Auxiliary files
│   └── 0002-another-post/
│       ├── .meta.json
│       └── post.md
├── .gitignore               # Auto-updated for private posts
└── README.md
```

## Example Workflow

```bash
# Initialize your blog
mkdir my-tech-blog && cd my-tech-blog
gblog init

# Create a new post
gblog new
# Interactive prompts:
# Title: "Getting Started with Go Generics"
# Description: "A practical guide to using generics in Go"
# Public: y

# Edit the post
gblog edit 0001
# Add content to posts/0001-getting-started-with-go-generics/post.md
# Add example files like generics-example.go

# List your posts
gblog list

# Publish when ready
gblog publish 0001
# Opens gist in browser: https://gist.github.com/yourusername/...

# Export all posts for backup
gblog export my-blog-backup.zip
```

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
# Clone the repository
git clone https://github.com/your-username/gblog
cd gblog

# Install dependencies
go mod tidy

# Build
make build

# Run during development
make run ARGS="list"

# Create new post during development
make new

# Publish post during development
make publish POST_ID=0001
```

## Privacy & Security

- **Private posts** are automatically added to `.gitignore`
- **Public posts** are committed to your repository and published as public gists
- **GitHub authentication** is handled by GitHub CLI
