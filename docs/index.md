# Box - Minimalist project-local toolbox

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

Download the latest binary for your platform from the [Releases](https://github.com/sebakri/etc/releases) page and place it in your system PATH.

### 2. Configure Your Project

Create a `box.yml` in your project root:

```yaml
tools:
  - name: task
    type: go
    source: github.com/go-task/task/v3/cmd/task@latest
  - name: ruff
    type: uv
    source: ruff
  - name: jj
    type: cargo
    source: jj-cli
    args:
      - --strategies
      - crate-meta-data
  - name: colorize
    type: gem
    source: colorize
env:
  DEBUG: "true"
  API_URL: "http://localhost:8080"
```

### 3. Install Tools

Run the install command to fetch and install all defined tools:

```bash
box install
```

### 4. Setup Shell Integration (Optional)

If you use `direnv`, generate the `.envrc` file:

```bash
box generate direnv
```

### 5. Use Your Tools

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
- `box run <command>`: Executes a binary from the local `.box/bin` directory.
- `box env`: Displays the merged list of environment variables.
- `box generate direnv`: Generates a `.envrc` file for `direnv` integration.
- `box doctor`: Checks if the required host runtimes (Go, npm, Cargo, uv) are installed.
- `box help`: Displays usage information.

## Contributing

We welcome contributions! Please check the [repository](https://github.com/sebakri/etc) for issues and pull requests.
