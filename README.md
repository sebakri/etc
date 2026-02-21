# Box - Minimalist project-local toolbox

[![Deploy Documentation](https://github.com/sebakri/box/actions/workflows/docs.yml/badge.svg)](https://sebakri.github.io/box/)
[![Release](https://github.com/sebakri/box/actions/workflows/release.yml/badge.svg)](https://github.com/sebakri/box/releases)

Box is a minimalist, project-local toolbox that keeps your development tools, binaries, and environment variables neatly packed and isolated within your project. It ensures a reproducible environment without cluttering your global system.

## Why use Box?

While there are many tool managers (like `asdf`, `mise`, or `aqua`), Box is designed for simplicity and project-level isolation:

- **No Magic/No Registry**: Box doesn't need a central database of "supported" tools. If it can be installed via `go install`, `npm`, `cargo`, `uv`, or a shell script, Box can manage it.
- **True Project Isolation**: Everything—binaries, caches, and metadata—lives inside your project's `.box` folder. Deleting the folder completely removes the tools.
- **Zero Dependencies**: Box is a single Go binary. You don't need Nix, a plugin system, or a complex runtime to get started.
- **Transparent Wrapper**: It doesn't replace your package managers; it coordinates them to keep your workspace clean.
- **Secure-by-Default**: Custom scripts and tools run in a restricted sandbox, preventing them from damaging your system outside the project directory.

## Documentation

Full documentation is available at [https://sebakri.github.io/box/](https://sebakri.github.io/box/)

## Quick Start

1.  **Configure**: Create a `box.yml` in your project root:
    ```yaml
    tools:
      - type: go
        source: github.com/go-task/task/v3/cmd/task
        version: latest
        binaries:
          - task
      - type: cargo
        source: jj-cli
        args:
          - --strategies
          - crate-meta-data
      - type: script
        alias: golangci-lint
        source:
          - curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $BOX_BIN_DIR v1.60.1
    env:
      APP_DEBUG: "true"
    ```
2.  **Install**: Run `box install`.
3.  **Setup Shell (Optional)**: Run `box generate direnv` if you use `direnv`.
4.  **Run**: Run `box run <tool>` or use `direnv`.

## Features

- **Project-Local Tools**: Installs tools into a local `.box/bin` directory.
- **Environment Variables**: Define project-specific environment variables in `box.yml`. `BOX_DIR`, `BOX_BIN_DIR`, `BOX_OS`, and `BOX_ARCH` are automatically provided.
- **No Root Required**: Leverages user-space package managers (Go, npm, Cargo, uv, gem) or custom shell scripts.
- **Declarative Configuration**: Defined in `box.yml`.
- **Manual or Automatic PATH**: Use `box run` or generate a `.envrc` for `direnv`.
- **Docker Integration**: Generate a pre-configured `Dockerfile` with all your tools.
- **Mandatory Sandboxing**: Custom scripts and tools are automatically isolated on macOS and Linux.

## Security

Box takes security seriously by implementing several layers of protection:

- **Mandatory Sandboxing**: Any `script` defined in `box.yml` and any tool executed via `box run` is automatically sandboxed using OS-native primitives (`sandbox-exec` on macOS, User Namespaces on Linux).
- **Strict Isolation**: Sandboxed scripts are restricted to writing only within the project root, the `.box` directory, and a dedicated session-specific temporary directory.
- **Environment Protection**: All environment variables and paths exported to `.envrc` are properly escaped to prevent shell injection attacks.
- **Transparent Manifest**: Tool tracking is stored in a human-readable JSON format (`.box/manifest.json`), allowing you to audit exactly what was installed.

## Installation

Install using curl:

```bash
curl -fsSL https://raw.githubusercontent.com/sebakri/box/main/scripts/install.sh | sh
```

Or download the binary for your platform from the [latest releases](https://github.com/sebakri/box/releases).

## Commands

- `box install [-y] [-f file]`: Installs tools defined in `box.yml` (or specified file). Use `-y` for non-interactive mode.
- `box list`: Lists installed tools and their binaries.
- `box run <command> [args...]`: Executes a binary from the local `.box/bin` directory.
- `box env [key]`: Displays the merged list of environment variables, or just the value of `key`.
- `box generate direnv`: Generates a `.envrc` file for `direnv` integration.
- `box generate dockerfile`: Generates a `Dockerfile` for containerized development.
- `box doctor`: Checks if the host runtimes (Go, npm, Cargo, uv, gem) are installed.
- `box version`: Prints the current version of box.

## Development

Build with Task:

```bash
task build
```

Run tests:

```bash
task test
```
