# gydnc - Content-Addressable Guidance for AI Agents

`gydnc` (pronounced "guidance") is a command-line tool for managing structured guidance entities for AI agents. It provides a simple, Git-friendly way to create, organize, and retrieve guidance documents that can be used to instruct AI systems on how to perform tasks according to your requirements.

## Features

- **Simple, Composable Commands**: Create, retrieve, update, and delete guidance with a clean CLI interface
- **Git-Friendly Storage**: Store guidance as human-readable Markdown files with YAML frontmatter
- **Hierarchical Organization**: Organize guidance in logical hierarchies with aliases (e.g., `must/safety-first`)
- **Tag-Based Discovery**: Find relevant guidance through tag filtering
- **Multiple Backend Support**: Extensible storage backend architecture (filesystem storage implemented)
- **Integration-Test Friendly**: Comprehensive test harness for CLI operations

## Installation

### Quick Install (Recommended)

**Linux/macOS (one-liner with automatic fallback):**
```bash
curl -sSL https://raw.githubusercontent.com/ofthemachine/gydnc/main/install.sh | sh
```
*This will try to download a pre-built binary, or automatically build from source if needed.*

**Manual Download (when pre-built binaries are available):**
```bash
# Detect platform and download latest release
PLATFORM=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')
curl -L -o gydnc "https://github.com/ofthemachine/gydnc/releases/latest/download/gydnc-${PLATFORM}-${ARCH}"
chmod +x gydnc && sudo mv gydnc /usr/local/bin/
```

**Platform-specific (when pre-built binaries are available):**
```bash
# Linux (x86_64)
curl -L -o gydnc https://github.com/ofthemachine/gydnc/releases/latest/download/gydnc-linux-amd64
chmod +x gydnc && sudo mv gydnc /usr/local/bin/

# macOS (Intel)
curl -L -o gydnc https://github.com/ofthemachine/gydnc/releases/latest/download/gydnc-darwin-amd64
chmod +x gydnc && sudo mv gydnc /usr/local/bin/

# macOS (Apple Silicon)
curl -L -o gydnc https://github.com/ofthemachine/gydnc/releases/latest/download/gydnc-darwin-arm64
chmod +x gydnc && sudo mv gydnc /usr/local/bin/
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/ofthemachine/gydnc.git

# Build gydnc
cd gydnc && make build

# Move the binary to your PATH
sudo mv gydnc /usr/local/bin/
```

## Getting Started

1. **Initialize gydnc** in a Git repository (strongly recommended for version control):

```bash
mkdir my-guidance && cd my-guidance
git init  # Create a Git repository first
gydnc init .  # Initialize gydnc in this Git repository
```

2. **Set up your environment**:

```bash
# Add this to your .bashrc, .zshrc, or similar shell configuration file
export GYDNC_CONFIG="/path/to/your/my-guidance/.gydnc/config.yml"
```

3. **Create your first guidance entity**:

```bash
# Create a new guidance entity with title, description, and tags
gydnc create --title "Safety First" \
    --description "Guidelines for ensuring code safety" \
    --tags quality:safety,scope:universal \
    --body "# Safety Guidelines\n\nAlways validate user input.\n" \
    must/safety-first
```

4. **List available guidance**:

```bash
# List all guidance entities
gydnc list

# Filter by tags
gydnc list --filter "tags:quality:safety"
```

5. **Retrieve guidance**:

```bash
# Get a specific guidance entity
gydnc get must/safety-first

# Get multiple guidance entities at once
gydnc get must/safety-first recipes/git/commit-creation
```

6. **Update existing guidance**:

```bash
# Update metadata (title, description)
gydnc update must/safety-first --title "Updated Title" --description "New description"

# Update tags (add or remove)
gydnc update must/safety-first --add-tag "quality:critical,priority:high" --remove-tag "scope:universal"

# Update content body by piping new content
cat updated_content.md | gydnc update must/safety-first
```

## Usage with AI Assistants

gydnc is designed to work seamlessly with AI assistants. When working with an AI, use the following workflow:

1. Start by getting an overview of available guidance:
   ```bash
   gydnc list
   ```

2. Retrieve relevant guidance based on your current task:
   ```bash
   gydnc get <entity1> <entity2>
   ```

3. As your conversation evolves, fetch additional guidance as needed.

## Architecture

gydnc uses a service-oriented architecture with:

- Command layer (CLI interface)
- Service layer (business logic)
- Storage layer (backend implementations)

Guidance entities are stored as `.g6e` files with YAML frontmatter containing metadata (title, description, tags) and Markdown body content.

## Testing

gydnc includes a comprehensive integration test framework that uses a declarative approach to testing CLI behavior:

```bash
# Run integration tests
make test-integration
```
