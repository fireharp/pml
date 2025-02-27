# PML - Programming with Machine Learning

## Introduction

PML (Programming with Machine Learning) is a programming paradigm that seamlessly integrates LLM capabilities into your development workflow through special directives. PML allows you to write code that interacts with LLMs using a simple and intuitive syntax, making AI-assisted programming more accessible and structured.

- All LLM and related calls are wrapped in special `:` directives
- Makes code more interactive and easier to work with
- No need to think about complex agentic frameworks - just write Python code and think about `:` directives as async function calls
- PML evolves as technology advances - migrating to newer APIs and protocols, embracing the most recent AI technologies
- Directives create clear boundaries to control where LLMs can make changes in your code

## Quick Start

1. Create a file with a `.pml` extension
2. Add `:ask` directives for LLM queries

```
:ask
What's the capital of France?
:--
```

When processed, this will be rendered as:

```
:ask
What's the capital of France?
:--(happy_panda:"Paris")
```

## Core Concepts

PML is built around these fundamental concepts:

1. **Directives**: Special syntax blocks starting with `:` that indicate LLM-related operations
2. **Interactive Execution**: Directives are processed at runtime and generate results
3. **Integration with Python**: PML works alongside standard Python code
4. **State Management**: PML maintains state across multiple queries and operations
5. **Boundary Control**: Directives create strict encapsulation boundaries, ensuring LLMs only modify code within designated areas

## Installation

See the [Getting Started](getting-started.md) guide for installation instructions.

## Project Status

PML is currently in active development. The implementation (impl1) focuses on core functionality with planned enhancements for improved directive support and integration capabilities.
