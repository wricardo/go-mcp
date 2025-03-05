# godoc-mcp

[![Go Report Card](https://goreportcard.com/badge/github.com/mrjoshuak/godoc-mcp)](https://goreportcard.com/report/github.com/mrjoshuak/godoc-mcp)
[![GoDoc](https://godoc.org/github.com/mrjoshuak/godoc-mcp?status.svg)](https://godoc.org/github.com/mrjoshuak/godoc-mcp)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Overview

`godoc-mcp` is a Model Context Protocol (MCP) server that provides efficient access to Go documentation. It helps LLMs understand Go projects by providing direct access to package documentation without needing to read entire source files. `godoc-mcp` can vastly improve the performance of using LLMs to develop in Go by substantially reducing the number of tokens needed to understand and make use of Go packages.

## Getting Started

```bash
go install github.com/mrjoshuak/godoc-mcp@latest
```

## Why Use godoc-mcp?

In a sentence: **`godoc-mcp` provides a more token efficient way for LLMs to understand Go projects.**

Traditional file-reading approaches require LLMs to process entire source files often many files to understand a single package. `godoc-mcp` provides several advantages:

1. **Token Efficiency**: Returns only the essential documentation, reducing token usage significantly
2. **Structured Information**: Provides official package documentation in a consistent, well-structured format
3. **Project Navigation**: Smart handling of project structures helps LLMs understand multi-package projects
4. **Integration Ready**: Works alongside other MCP servers, enabling both high-level and detailed code analysis
5. **Performance**: Caching and optimized token usage make `godoc-mcp` a fast and efficient tool for Go development
6. **Local**: Does not require an internet connection to access documentation

With `godoc-mcp`, a LLM can get precisely the information it needs without having to read entire source files. Here are the different levels of detail that an LLM can get.

- Documentation for one exported symbol
- The complete source for one symbol
- A list of all exported symbols (the concise documentation)
- A list of all symbols including unexported symbols
- The full documentation for a package
- The entire source for a package

This makes `godoc-mcp` an essential tool for Go developers using LLMs by enabling LLMs to understand significantly more, and in more detail, about the context than previously possible in any programming language.

## Features

The server will:
1. For directories with Go files: Return package documentation
2. For directories without Go files: List available Go packages in subdirectories
3. For import paths: Return standard library or third-party package documentation

- **Efficient Documentation Access**: Retrieves official Go documentation with minimal token usage
- **Smart Package Discovery**: When pointed at a directory without Go files, lists available Go packages in subdirectories
- **Flexible Path Support**:
  - Local file paths (e.g., "/full/path/to/mypackage")
  - Import paths (e.g., "io", "github.com/user/repo")
- **Automatic Module Context**:
  - Creates temporary Go projects when needed
  - Automatically sets up module context for external packages
  - No manual module setup required for any package documentation
  - Handles cleanup of temporary projects
- **Module-Aware**: Supports documentation for third-party packages through working directory context (i.e. it will run `go doc` from the working directory)
- **Performance Optimized**:
  - Built-in response caching
  - Efficient token usage through focused documentation retrieval
  - Metadata about response sizes
  - Smart handling of standard library vs external packages

### Examples

In addition to providing documentation while working on coding tasks. `godoc-mcp` can also be used to explore Go projects and packages. Here are some examples for general prompting:

#### Project Understanding

"I'm looking at a Go project at /path/to/some/project. What packages does it contain and what do they do?"

#### Package Interface Understanding

"What interfaces does the io package provide? I'm particularly interested in anything related to reading."

#### Implementation Guidance

"I need to implement the io.Reader interface. Show me its documentation and any related types I should know about."

#### API Usage

"Show me the documentation for the Resource type in the /path/to/some/project. I need to understand how to create and use it."

#### Library Exploration

"I'm in /path/to/some/project which uses github.com/gorilla/mux. Show me the documentation for the Router type."

#### Method Discovery

"What methods are available on the http.Request type? I'm working with standard library HTTP handlers."

#### Focused Learning

"Explain how to configure the Server type in the /path/to/project/server package."

#### Package Browsing

"I'm in a new Go project directory and see multiple packages. Can you show me what each one does?"

## Usage

To add to the Claude desktop app:

```yaml
{
  "mcpServers": {
    # other MCP servers ...
    "godoc": {
      "command": "/path/to/godoc-mcp",
      "args": [],
      "env": {
        "GOPATH": "/path/to/go",
        "GOMODCACHE": "/path/to/go/pkg/mod"
      }
    }
  }
}
```

When connected to an MCP-capable LLM (like Claude), godoc-mcp provides the `get_doc` tool with the following parameters:

- `path`: Path to the Go package or file (import path or file path)
- `target` (optional): Specific symbol to document (function, type, etc.)
- `cmd_flags` (optional): Additional go doc command flags
- `working_dir` (optional): Working directory for module-aware documentation (if not provided, a temporary project will be created automatically)

Advanced `cmd_flags` values that an LLM can leverage:
- `-all`: Show all documentation for package, excluding unexported symbols
- `-u`: Show unexported symbols
- `-src`: Show the source code instead of documentation

## Troubleshooting

- For local paths, ensure they contain Go source files or point to directories containing Go packages
- If you see module-related errors, ensure GOPATH and GOMODCACHE environment variables are set correctly in your MCP server configuration
- The server automatically handles module context for external packages, but you can still provide a specific working_dir if needed for special cases

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
