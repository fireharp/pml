# PML (Programming with Machine Learning) - Implementation 1

A tool for integrating LLM-powered code generation and processing into your development workflow through special PML files.

## Quick Start (Using Task)

1. Install [Task](https://taskfile.dev) if you haven't already
2. Build the binary:
   ```bash
   task build
   ```
3. Or build for all platforms:
   ```bash
   task build:all
   ```
   This will create binaries for:
   - Darwin (macOS) AMD64: `bin/pml-watcher-darwin-amd64`
   - Darwin (macOS) ARM64: `bin/pml-watcher-darwin-arm64`
   - Linux AMD64: `bin/pml-watcher-linux`
   - Windows AMD64: `bin/pml-watcher-windows.exe`

### Available Task Commands

```bash
task --list         # Show all available commands
task build          # Build for current platform
task build:all      # Build for all platforms
task clean          # Clean build artifacts
task test           # Run tests
task test:watch     # Run tests in watch mode
```

### Using the Compiled Binary

Instead of `go run main.go`, you can use the compiled binary:

```bash
./bin/pml-watcher -file path/to/your/file.pml
./bin/pml-watcher -force
./bin/pml-watcher -cleanup
```

## Overview

PML allows you to write files with special blocks that get processed by an LLM (Language Learning Model). These blocks can contain questions, code generation requests, or other prompts that the LLM will process and respond to.

## Installation

1. Clone the repository
2. Set up your environment variables in a `.env` file:
   ```
   OPENAI_API_KEY=your_api_key_here
   PML_DEBUG=1  # Optional: Enable debug logging
   ```

## Directory Structure

The tool expects/creates the following directory structure in your workspace:

```
your-workspace/
├── sources/     # Directory containing your .pml files
└── results/     # Directory where processed results are stored
```

## File Format

PML files (`.pml` extension) can contain special blocks marked with `:ask` and `:--`:

```
:ask
What is the best way to implement a binary search tree in Go?
:--

:ask
Generate a unit test for the following function:
func Add(a, b int) int {
    return a + b
}
:--
```

## Usage

The tool provides several command-line options for processing PML files:

### Process a Single File

```bash
go run main.go -file path/to/your/file.pml
```

### Process All Files

```bash
go run main.go
```

This will process all `.pml` files in the `sources` directory.

### Force Processing

To force processing of files, ignoring any cache:

```bash
go run main.go -force
```

### Cleanup Generated Files

To remove all generated files (`.pml.py` files and `.pml` directories):

```bash
go run main.go -cleanup
```

## Command Line Options

- `-file string`: Process only a specific file
- `-force`: Force processing of all files, ignoring cache
- `-cleanup`: Clean up all generated files

## Example

1. Create a file `sources/example.pml`:

   ```
   :ask
   Write a Go function that calculates the Fibonacci sequence.
   :--
   ```

2. Process the file:

   ```bash
   go run main.go -file sources/example.pml
   ```

3. The LLM will process the request and generate a response, which will be integrated into your workflow.

## Notes

- The tool creates necessary directories automatically
- Generated files are stored with `.pml.py` extension
- Use the cleanup option to remove all generated files when needed
- Debug logging can be enabled by setting `PML_DEBUG=1` in your environment
