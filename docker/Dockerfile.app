FROM --platform=$BUILDPLATFORM ubuntu:22.04 AS builder

# Get target platform architecture
ARG TARGETARCH
ARG TARGETOS

# Install basic tools
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    build-essential \
    wget \
    ca-certificates \
    clang \
    iptables \
    tcpdump \
    ipset \
    git \
    && rm -rf /var/lib/apt/lists/*

# Set Go version
ENV GO_VERSION=1.21.2

# Set Go download URL based on architecture
RUN case "${TARGETARCH}" in \
    "amd64") \
        GO_ARCH="linux-amd64" \
        ;; \
    "arm64") \
        GO_ARCH="linux-arm64" \
        ;; \
    "arm") \
        GO_ARCH="linux-armv6l" \
        ;; \
    *) \
        echo "Unsupported architecture: ${TARGETARCH}" && exit 1 \
        ;; \
    esac && \
    wget https://golang.org/dl/go${GO_VERSION}.${GO_ARCH}.tar.gz -O /tmp/go.tar.gz && \
    tar -C /usr/local -xzf /tmp/go.tar.gz && \
    rm /tmp/go.tar.gz

# Set Go environment variables
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH=/go
ENV PATH="${GOPATH}/bin:${PATH}"
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}
ENV CGO_ENABLED=1

# Verify installations
RUN go version && \
    gcc --version && \
    make --version
# Set working directory
WORKDIR /app

# Copy the source code
COPY ./web-app .
##
# Build the application
RUN CGO_ENABLED=0 GOOS=linux go mod tidy && go build -o app

# Stage 2: Create a minimal runtime image
FROM ubuntu:22.04
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    build-essential \
    wget \
    ca-certificates \
    clang \
    iptables \
    tcpdump \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/app /app
COPY ./web-app/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Expose port 8080
EXPOSE 8080

# Command to run the application
ENTRYPOINT ["/entrypoint.sh"]
