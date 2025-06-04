.PHONY: build install clean run new publish list edit export help

# Build the application
build:
	@echo "🔨 Building gblog..."
	@go build -o bin/gblog

# Install the application to GOPATH/bin
install:
	@echo "📦 Installing gblog..."
	@go install

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	@rm -rf bin/

# Run the application (for development)
run:
	@go run main.go $(ARGS)

# Create a new post (interactive)
new:
	@go run main.go new

# Publish a post (requires POST_ID)
publish:
	@if [ -z "$(POST_ID)" ]; then \
		echo "❌ Please specify POST_ID: make publish POST_ID=0001"; \
		exit 1; \
	fi
	@go run main.go publish $(POST_ID)

# List all posts
list:
	@go run main.go list

# Edit a post (requires POST_ID)
edit:
	@if [ -z "$(POST_ID)" ]; then \
		echo "❌ Please specify POST_ID: make edit POST_ID=0001"; \
		exit 1; \
	fi
	@go run main.go edit $(POST_ID)

# Export all posts
export:
	@go run main.go export

# Show help
help:
	@echo "📝 gblog - Gist-powered blog CLI"
	@echo ""
	@echo "Available targets:"
	@echo "  build     - Build the application"
	@echo "  install   - Install to GOPATH/bin"
	@echo "  clean     - Clean build artifacts"
	@echo "  run       - Run the application (use ARGS for arguments)"
	@echo "  new       - Create a new post interactively"
	@echo "  publish   - Publish a post (use POST_ID=xxxx)"
	@echo "  list      - List all posts"
	@echo "  edit      - Edit a post (use POST_ID=xxxx)"
	@echo "  export    - Export all posts to zip"
	@echo "  help      - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make new"
	@echo "  make publish POST_ID=0001"
	@echo "  make edit POST_ID=0001"
	@echo "  make list"
	@echo "  make export"

# Default target
all: build
