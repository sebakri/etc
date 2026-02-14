---
hide:
  - navigation
  - toc
---

# Box - Minimalist project-local toolbox

![Box Logo](assets/logo.jpg){: style="width: 50%; display: block; margin: 0 auto;" }

Box is a minimalist, project-local toolbox that keeps your development tools, binaries, and environment variables neatly packed and isolated within your project. It allows you to define and install project-specific tools locally without requiring root permissions, keeping your global environment clean and your development setup reproducible.

## Key Features

- **Project-Local Tools**: Tools are installed in `.box/bin`, isolated from your system.
- **Environment Variables**: Define project-specific variables that are automatically exported.
- **Declarative Setup**: Define all required tools and env vars in a simple `box.yml` file.
- **Multi-Runtime Support**: Works seamlessly with Go, npm, Cargo, uv, and gem.
- **direnv Integration**: Automatically manages your `PATH` and `ENV` using `.envrc`.
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

### 3. Install Tools

Run the install command to fetch and install all defined tools. Custom scripts have access to several environment variables:

- `BOX_DIR`: The path to the `.box` directory.
- `BOX_BIN_DIR`: The path to the `.box/bin` directory.
- `BOX_OS`: The current operating system (`runtime.GOOS`).
- `BOX_ARCH`: The current architecture (`runtime.GOARCH`).

`.box/bin` is also added to their `PATH`.

```bash
box install
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

- `box install`: Installs tools defined in `box.yml`.
- `box list`: Lists installed tools and their binaries.
- `box run <command>`: Executes a binary from the local `.box/bin` directory.
- `box env`: Displays the merged list of environment variables.
- `box generate direnv`: Generates a `.envrc` file for `direnv` integration.
- `box generate dockerfile`: Generates a `Dockerfile` for containerized development.
- `box doctor`: Checks if the required host runtimes (Go, npm, Cargo, uv) are installed.
- `box help`: Displays usage information.

## Contributing

We welcome contributions! Please check the [repository](https://github.com/sebakri/box) for issues and pull requests.
