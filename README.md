# gh-issue-dependency

A GitHub CLI extension for managing issue dependencies and sub-issues, enabling better project organization and tracking of complex tasks.

## Features

- Create and manage sub-issues with automatic dependency tracking
- Visualize issue dependency graphs
- Organize complex projects with hierarchical issue structures
- Seamless integration with GitHub Issues and Projects

## Installation

### Prerequisites

- Go 1.19 or later
- GitHub CLI (`gh`) installed and authenticated

### Install from source

```bash
git clone https://github.com/torynet/gh-issue-dependency
cd gh-issue-dependency
go build -o gh-issue-dependency
```

### Install as GitHub CLI extension

```bash
gh extension install torynet/gh-issue-dependency
```

## Quick Start

```bash
# Create a sub-issue
gh issue-dependency create "Implement user authentication" --parent 123

# List dependencies for an issue
gh issue-dependency list 123

# Show dependency graph
gh issue-dependency graph
```

## Development

### Building from source

```bash
go build -o gh-issue-dependency ./main.go
```

### Running tests

```bash
go test ./...
```

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Project Structure

```
gh-issue-dependency/
├── main.go                      # Application entry point
├── cmd/                         # Command implementations
├── pkg/                         # Shared utilities and types  
├── tests/                       # Integration tests
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── README.md                    # This file
├── LICENSE                      # MIT license
└── .gitignore                   # Git ignore patterns
```
