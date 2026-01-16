FROM debian:bookworm-slim

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl ca-certificates git build-essential \
    nodejs npm cargo ruby-full \
    && rm -rf /var/lib/apt/lists/*

# Install latest Go
RUN curl -LsSf https://go.dev/dl/go1.24.0.linux-amd64.tar.gz | tar -C /usr/local -xz
ENV PATH="/usr/local/go/bin:${PATH}"

# Install uv globally (accessible by all users)
ENV UV_INSTALL_DIR="/usr/local/bin"
RUN curl -LsSf https://astral.sh/uv/install.sh | sh

# Copy box binary
COPY --link --chmod=755 box /usr/local/bin/box

# Set up user and workspace
RUN useradd -m -s /bin/bash box
USER box
WORKDIR /home/box

# Copy configuration and install tools
COPY --chown=box:box box.yml .
ENV CGO_ENABLED=0
RUN box install --non-interactive

# Add box binaries to PATH
ENV PATH="/home/box/.box/bin:${PATH}"

ENTRYPOINT ["/bin/bash"]
