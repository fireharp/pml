# PML Roadmap

This document outlines the planned development trajectory for PML, highlighting upcoming features and improvements.

## Current Status

The current implementation (impl1) provides:

- Basic `:ask` directive functionality
- Simple `:do` directive support
- File-based processing system
- OpenAI LLM integration

## Short-Term Priorities

### 1. Parser Enhancements

- **Decomposition**: Break down the parser into smaller, more focused files
- **Error Handling**: Improve error reporting and handling
- **Performance**: Optimize parsing for larger files

### 2. Directive Expansion

- **Enhanced `:ask`**: Support for named blocks and return types
- **Enhanced `:do`**: More declarative syntax for workflow control
- **New `:context`**: Implementation of context management for state persistence

### 3. Processing Improvements

- **Asynchronous Block Processing**: Launch blocks asynchronously instead of sequentially
- **Block Execution Isolation**: Keep block execution code in dedicated folders
- **Cleanup Improvements**: Better management of temporary files

### 4. Metrics and Monitoring

- **Performance Tracking**: Add metrics for LLM calls (time-to-first-token, total time)
- **Integration with Observability Tools**: Add support for tracing and monitoring tools
- **Debug Logging**: Enhanced logging for troubleshooting

## Medium-Term Goals

### 1. Python Integration

- **Python API**: First-class Python library for PML
- **IDE Integration**: Support for major IDEs (VS Code, JetBrains)
- **Type Checking**: Better integration with Python type systems

### 2. Advanced Directives

- **`:reflect` Implementation**: Quality control and validation mechanisms
- **`:agent` Directive**: Support for agent-based workflows
- **Tool Integration**: Allow directives to interact with external tools

### 3. System Improvements

- **Caching Strategy**: More intelligent caching of LLM responses
- **Distribution**: Improved packaging and distribution
- **Documentation**: Comprehensive documentation and examples

## Long-Term Vision

### 1. Ecosystem Development

- **Plugin System**: Allow for community-contributed extensions
- **Template Library**: Pre-built PML templates for common tasks
- **Multi-LLM Support**: First-class support for multiple LLM providers

### 2. Advanced Features

- **Session Management**: Persistent sessions across multiple runs
- **Collaborative Workflows**: Support for multi-user environments
- **Custom Runtime**: Specialized runtime for PML files

### 3. Domain-Specific Extensions

- **Data Science Kit**: Specialized directives for data analysis workflows
- **Web Development Kit**: Tools for web application development
- **Documentation Kit**: Enhanced support for documentation generation

## Contributing

The PML roadmap is open to community input and contributions. If you have suggestions or would like to contribute to any of these initiatives, please:

1. Open an issue in the GitHub repository
2. Join the discussion in our community channels
3. Submit pull requests for implemented features

We welcome feedback on prioritization and feature requests that would make PML more useful for your specific use cases.
