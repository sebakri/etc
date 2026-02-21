---
hide:
  - navigation
  - toc
---

# Box - Minimalist project-local toolbox

![Box Logo](assets/logo.jpg){: style="width: 50%; display: block; margin: 0 auto;" }

Box is a minimalist, project-local toolbox that keeps your development tools, binaries, and environment variables neatly packed and isolated within your project. It allows you to define and install project-specific tools locally without requiring root permissions, keeping your global environment clean and your development setup reproducible.

## Why use Box?

In an ecosystem with many tool managers, Box stands out for its "glue" philosophy:

- **Minimalist & Agnostic**: Box is a thin wrapper. It doesn't require a central registry or custom plugins for every tool. If you know how to install it via a standard package manager (Go, Cargo, npm, uv, Gem), Box can automate it.
- **Portable "Toolbox"**: By keeping everything in `.box/`, your project becomes a self-contained unit. You can move the project folder, and your tools move with it.
- **Scriptable**: The `script` installer type allows you to handle edge cases and proprietary tools using standard shell commands, while still benefiting from Box's environment management.
- **CI/CD Ready**: Since it's a single binary and uses standard installers, it's trivial to use in GitHub Actions or any other CI provider to ensure your build environment matches your local one.
- **Secure-by-Default**: Custom scripts and tools run in a restricted sandbox, preventing them from damaging your system outside the project directory.

## Key Features

- **Project-Local Tools**: Tools are installed in `.box/bin`, isolated from your system.
- **Environment Variables**: Define project-specific variables that are automatically exported.
- **Declarative Setup**: Define all required tools and env vars in a simple `box.yml` file.
- **Multi-Runtime Support**: Works seamlessly with Go, npm, Cargo, uv, and gem.
- **direnv Integration**: Automatically manages your `PATH` and `ENV` using `.envrc`.
- **Mandatory Sandboxing**: Custom scripts and tools are automatically isolated on macOS and Linux.
- **Cross-Platform**: Built in Go, supporting Linux, macOS, and Windows.

## Quick Start

### 1. Installation

Download the latest binary for your platform from the [Releases](https://github.com/sebakri/box/releases) page and place it in your system PATH.

### Installer Script (Linux & macOS)

Alternatively, you can use the installer script to automatically download and install the latest version of `box`:

```bash
curl -sSfL https://raw.githubusercontent.com/sebakri/box/main/scripts/install.sh | sh
```

The script will detect your OS and architecture, download the appropriate binary, and install it into `/usr/local/bin`. If you want to install it in a different directory, you can set the `BOX_INSTALL_DIR` environment variable:

```bash
export BOX_INSTALL_DIR="$HOME/.local/bin"
curl -sSfL https://raw.githubusercontent.com/sebakri/box/main/scripts/install.sh | sh
```

### 2. Configure Your Project

Create a `box.yml` in your project root:

```yaml
tools:
  - type: go
    source: golang.org/x/tools/gopls
    version: latest
  - type: cargo
    source: jj-cli
    args:
      - --strategies
      - crate-meta-data
  - type: go
    source: github.com/go-task/task/v3/cmd/task
    version: latest
  - type: uv
    source: mkdocs
    args:
      - --with
      - mkdocs-simple-blog
  - type: uv
    source: rich-cli
  - type: uv
    source: black
    version: 24.3.0
  - type: script
    alias: golangci-lint with script
    source:
      - echo "Installing golangci-lint on $BOX_OS ($BOX_ARCH)"
      - sleep 5 # So you can read the message
      - curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $BOX_BIN_DIR v1.60.1
  - type: npm
    source: cowsay
```

## Configuration Reference (`box.yml`)

The `box.yml` file supports the following fields for each tool:

- `type`: The installer to use (`go`, `npm`, `cargo`, `uv`, `gem`, `script`).
- `source`: The package name, URL, or script commands.
- `version`: (Optional) The version to install.
- `alias`: (Optional) A human-readable name for the tool.
- `args`: (Optional) Additional arguments passed to the underlying installer.
- `binaries`: (Optional) Explicit list of binary names to link into `.box/bin`. Useful for Go tools that install multiple binaries or if the default name detection fails.

### The `script` Installer

The `script` type allows you to install tools that don't have a supported package manager. It is **not** for general-purpose hooks, but for running custom installation logic.

Scripts have access to several environment variables:

- `BOX_DIR`: The path to the `.box` directory.
- `BOX_BIN_DIR`: The path to the `.box/bin` directory.
- `BOX_OS`: The current operating system (`runtime.GOOS`).
- `BOX_ARCH`: The current architecture (`runtime.GOARCH`).

## Security

Box takes security seriously by implementing several layers of protection:

- **Mandatory Sandboxing**: Any `script` defined in `box.yml` and any tool executed via `box run` is automatically sandboxed using OS-native primitives (`sandbox-exec` on macOS, User Namespaces on Linux).
- **Strict Isolation**: Sandboxed scripts are restricted to writing only within the project root, the `.box` directory, and a dedicated session-specific temporary directory.
- **Environment Protection**: All environment variables and paths exported to `.envrc` are properly escaped to prevent shell injection attacks.
- **Transparent Manifest**: Tool tracking is stored in a human-readable JSON format (`.box/manifest.json`), allowing you to audit exactly what was installed.

### 3. Install Tools

Run the install command to fetch and install all defined tools.

```bash
box install [-y] [-f file]
```

### 4. Setup Shell Integration (Optional)

If you use `direnv`, generate the `.envrc` file:

```bash
box generate direnv
```

### 5. Docker Integration (Optional)

Box can generate a `Dockerfile` that sets up a development environment with all your tools pre-installed:

```bash
box generate dockerfile
```

This creates a `debian:bookworm-slim` based image with the necessary runtimes and your `box.yml` tools installed.

### 6. Use Your Tools

You can run tools directly using the `run` command:

```bash
box run task --version
```

Or, if you use `direnv`, simply use them directly:

```bash
task --version
```

## Commands

- `box install [-y] [-f file]`: Installs tools defined in `box.yml` (or specified file). Use `-y` or `--non-interactive` for CI environments.
- `box list`: Lists installed tools and their binaries.
- `box run <command> [args...]`: Executes a binary from the local `.box/bin` directory.
- `box env [key]`: Displays the merged list of environment variables. If `key` is provided, only its value is printed (useful for shell substitution like `$(box env BOX_DIR)`).
- `box generate direnv`: Generates a `.envrc` file for `direnv` integration.
- `box generate dockerfile`: Generates a `Dockerfile` for containerized development.
- `box doctor`: Checks if the required host runtimes (Go, npm, Cargo, uv, gem) are installed.
- `box version`: Prints the current version of box.
- `box help`: Displays usage information.

## Contributing

We welcome contributions! Please check the [repository](https://github.com/sebakri/box) for issues and pull requests.
