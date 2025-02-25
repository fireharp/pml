# PML Architecture

This document outlines the architecture of the current PML implementation (impl1). It provides an overview of the system components and how they interact.

## System Overview

The PML implementation consists of several core components:

1. **Parser System**: Processes `.pml` files and identifies special directive blocks
2. **LLM Integration**: Manages communication with language learning models
3. **File Processing**: Handles file I/O and updates

## Component Details

### Parser System

The parser is responsible for:

- Reading `.pml` files
- Identifying special directive blocks (`:ask`, `:do`, etc.)
- Processing directives based on their type
- Updating files with processing results

Key files:

- `parser/parser.go`: Core parsing logic

### LLM Integration

The LLM integration component:

- Provides a client interface to LLM services (primarily OpenAI)
- Manages API authentication and communication
- Formats prompts and processes responses
- Handles errors and retries

Key files:

- `llm/llm.go`: LLM client implementation

### Command-Line Interface

The CLI provides:

- File processing commands
- Forced processing options
- Cleanup functionality

Key files:

- `main.go`: Entry point and command processing

## Data Flow

The typical flow of data through the system is:

1. User creates or modifies a `.pml` file in the `sources` directory
2. The parser reads the file and identifies directive blocks
3. For each block:
   - The directive type is determined
   - Block content is extracted
   - For `:ask` blocks, the content is sent to the LLM
   - Results are captured
4. The file is updated with processing results
5. Processed results are stored in the `results` directory

## Future Architecture

Plans for evolving the architecture include:

1. **Module Decomposition**: Breaking down the parser into smaller, more focused components
2. **Asynchronous Processing**: Launching blocks asynchronously instead of sequentially
3. **Enhanced Debugging**: Adding more detailed logging and performance metrics
4. **Block Execution Isolation**: Keeping block execution code in dedicated folders

### Planned Components

- **Context Manager**: To handle state across multiple blocks
- **Tool Integration**: To allow directives to interact with external tools
- **Reflection System**: For validation and quality control
- **Enhanced LLM Metrics**: For tracking performance metrics like time-to-first-token

## Implementation Notes

- **Language**: The core implementation is in Go
- **Python Integration**: Some components interact with Python for specific functionality
- **Environment Configuration**: System parameters are configured via environment variables
- **Testing**: Test files cover the core components

## Directory Structure

```
impl1/
├── llm/        # LLM client integration
├── parser/     # PML file parsing
├── sources/    # Source PML files
├── results/    # Processed results
└── cmd/        # Command implementations
```

## Configuration

Configuration is primarily through environment variables:

- `OPENAI_API_KEY`: Required for LLM integration
- `PML_DEBUG`: Optional for enabling debug logging
