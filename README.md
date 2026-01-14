# Box - Minimalist project-local toolbox

[![Deploy Documentation](https://github.com/sebakri/etc/actions/workflows/docs.yml/badge.svg)](https://sebakri.github.io/etc/)
[![Release](https://github.com/sebakri/etc/actions/workflows/release.yml/badge.svg)](https://github.com/sebakri/etc/releases)

Box is a minimalist, project-local toolbox that keeps your development tools, binaries, and environment variables neatly packed and isolated within your project. It ensures a reproducible environment without cluttering your global system.

## Documentation

Full documentation is available at [https://sebakri.github.io/etc/](https://sebakri.github.io/etc/)

## Quick Start

1.  **Configure**: Create a `box.yml` in your project root:
    ```yaml
    tools:
      - name: task
        type: go
        source: github.com/go-task/task/v3/cmd/task@latest
      - name: jj
        type: cargo
        source: jj-cli
        args:
          - --strategies
          - crate-meta-data
    env:
      APP_DEBUG: "true"
    ```
2.  **Install**: Run `box install`.
3.  **Setup Shell (Optional)**: Run `box generate direnv` if you use `direnv`.
4.  **Run**: Run `box run <tool>` or use `direnv`.

## Features

- **Project-Local Tools**: Installs tools into a local `.box/bin` directory.
- **Environment Variables**: Define project-specific environment variables in `box.yml`.
- **No Root Required**: Leverages user-space package managers (Go, npm, Cargo, uv, gem).
- **Declarative Configuration**: Defined in `box.yml`.
- **Manual or Automatic PATH**: Use `box run` or generate a `.envrc` for `direnv`.

## Installation

Download the binary for your platform from the [latest releases](https://github.com/sebakri/etc/releases).

## Commands

- `box install`: Installs tools defined in `box.yml`.
- `box run <command>`: Executes a binary from the local `.box/bin` directory.
- `box env`: Displays the merged list of environment variables.
- `box generate direnv`: Generates a `.envrc` file for `direnv` integration.
- `box doctor`: Checks if the host runtimes are installed.

## Development

Build with Task:
```bash
task build
```

Run tests:
```bash
task test
```