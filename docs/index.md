# ETC - Ephemeral Tool Configuration

ETC is a lightweight, platform-independent tools manager for your projects. It allows you to define and install project-specific tools locally without requiring root permissions, keeping your global environment clean and your development setup reproducible.

## Key Features

- **Project-Local Tools**: Tools are installed in `.etc/bin`, isolated from your system.
- **Declarative Setup**: Define all required tools in a simple `etc.yml` file.
- **Multi-Runtime Support**: Works seamlessly with Go, npm, and Cargo.
- **direnv Integration**: Automatically manages your `PATH` using `.envrc`.
- **Cross-Platform**: Built in Go, supporting Linux, macOS, and Windows.

## Quick Start

### 1. Installation

Download the latest binary for your platform from the [Releases](https://github.com/sebakri/etc/releases) page and place it in your system PATH.

### 2. Configure Your Project

Create an `etc.yml` in your project root:

```yaml
tools:
  - name: task
    type: go
    source: github.com/go-task/task/v3/cmd/task@latest
  - name: prettier
    type: npm
    source: prettier
```

### 3. Install Tools

Run the install command to fetch and install all defined tools:

```bash
etc install
```

### 4. Use Your Tools

You can run tools directly using the `run` command:

```bash
etc run task --version
```

Or, if you use `direnv`, simply `allow` the `.envrc` and use them directly:

```bash
task --version
```

## Commands

- `etc install`: Installs tools defined in `etc.yml` and sets up `.envrc`.
- `etc run <command>`: Executes a binary from the local `.etc/bin` directory.
- `etc doctor`: Checks if the required host runtimes (Go, npm, Cargo) are installed.
- `etc help`: Displays usage information.

## Contributing

We welcome contributions! Please check the [repository](https://github.com/sebakri/etc) for issues and pull requests.
