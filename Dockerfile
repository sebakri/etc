FROM debian:bookworm-slim

# Package manager feature flags
ARG INSTALL_GO=true
ARG INSTALL_NODE=true
ARG INSTALL_CARGO=true
ARG INSTALL_UV=true
ARG INSTALL_RUBY=true

# Install system dependencies and selected package managers
RUN apt-get update && \
    PACKAGES="curl ca-certificates git build-essential direnv" && \
    if [ "$INSTALL_NODE" = "true" ]; then PACKAGES="$PACKAGES nodejs npm"; fi && \
    if [ "$INSTALL_RUBY" = "true" ]; then PACKAGES="$PACKAGES ruby-full"; fi && \
    apt-get install -y --no-install-recommends $PACKAGES && \
    rm -rf /var/lib/apt/lists/*

# Install latest Go if enabled
RUN if [ "$INSTALL_GO" = "true" ]; then \
    curl -LsSf https://go.dev/dl/go1.24.0.linux-amd64.tar.gz | tar -C /usr/local -xz; \
    fi
ENV PATH="/usr/local/go/bin:${PATH}"

# Install cargo-binstall if enabled
RUN if [ "$INSTALL_CARGO" = "true" ]; then \
    curl -L --proto '=https' --tlsv1.2 -sSf https://raw.githubusercontent.com/cargo-bins/cargo-binstall/main/install.sh | sh && \
    if [ -f "$HOME/.cargo/bin/cargo-binstall" ]; then mv "$HOME/.cargo/bin/cargo-binstall" /usr/local/bin/; fi; \
    fi

# Install uv globally if enabled
RUN if [ "$INSTALL_UV" = "true" ]; then \
    curl -LsSf https://astral.sh/uv/install.sh | UV_INSTALL_DIR=/usr/local/bin sh; \
    fi

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

CMD ["/bin/bash"]
