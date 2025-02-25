# Getting Started with PML

This guide will help you install and set up PML for your development environment.

## Prerequisites

- Go 1.19 or higher
- Python 3.8 or higher (for Python integration)
- OpenAI API key

## Installation

### Using Task (Recommended)

1. Install [Task](https://taskfile.dev) if you haven't already
2. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/pml.git
   cd pml
   ```
3. Build the binary:
   ```bash
   task build
   ```
4. Or build for all platforms:
   ```bash
   task build:all
   ```

This will create binaries for:

- Darwin (macOS) AMD64: `bin/pml-watcher-darwin-amd64`
- Darwin (macOS) ARM64: `bin/pml-watcher-darwin-arm64`
- Linux AMD64: `bin/pml-watcher-linux`
- Windows AMD64: `bin/pml-watcher-windows.exe`

### From Source

1. Clone the repository
2. Build the binary:
   ```bash
   go build -o bin/pml main.go
   ```

## Configuration

1. Create a `.env` file in the root directory:
   ```
   OPENAI_API_KEY=your_api_key_here
   PML_DEBUG=1  # Optional: Enable debug logging
   ```

## Directory Structure

PML works with the following directory structure:

```
your-workspace/
├── sources/     # Directory containing your .pml files
└── results/     # Directory where processed results are stored
```

These directories will be created automatically if they don't exist.

## Basic Usage

### Process a Single File

```bash
./bin/pml-watcher -file path/to/your/file.pml
```

### Process All Files

```bash
./bin/pml-watcher
```

This will process all `.pml` files in the `sources` directory.

### Force Processing

To force processing of files, ignoring any cache:

```bash
./bin/pml-watcher -force
```

### Cleanup Generated Files

To remove all generated files:

```bash
./bin/pml-watcher -cleanup
```

## Creating Your First PML File

1. Create a file `sources/example.pml`:

   ```
   :ask
   Write a Go function that calculates the Fibonacci sequence.
   :--
   ```

2. Process the file:

   ```bash
   ./bin/pml-watcher -file sources/example.pml
   ```

3. The LLM will process your request and generate a response in the file.

## Available Task Commands

```bash
task --list         # Show all available commands
task build          # Build for current platform
task build:all      # Build for all platforms
task clean          # Clean build artifacts
task test           # Run tests
task test:watch     # Run tests in watch mode
```

## Next Steps

Check out the [Directives](directives.md) documentation to learn about all available PML directives and their usage.
