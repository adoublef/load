# check=skip=SecretsUsedInArgOrEnv

# -------------------------------------------------------------
# Stage 1: Build binary & fetch keys (Hardened Go Dev image)
# -------------------------------------------------------------
FROM dhi.io/golang:1.27-dev AS builder

WORKDIR /src

RUN apt-get update && \
    apt-get install -y --no-install-recommends curl ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build static Go binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /bin/server .

# Fetch public keys for GitHub users at build time
ARG GITHUB_USERS="adoublef"
RUN touch /etc/authorized_keys && \
    for user in $(echo $GITHUB_USERS | tr "," "\n"); do \
        echo "Fetching keys for ${user}..."; \
        curl -fsSL "https://github.com/${user}.keys" >> /etc/authorized_keys; \
        echo "" >> /etc/authorized_keys; \
    done

# Prepare persistent SSH host key directory with non-root ownership
RUN mkdir -p /data/ssh && chown -R 65532:65532 /data/ssh /etc/authorized_keys

# -------------------------------------------------------------
# Stage 2: Docker Hardened Static Runtime (Zero-CVE, distroless)
# -------------------------------------------------------------
FROM dhi.io/static:20250419

WORKDIR /

# Copy only the compiled binary and the generated keys
COPY --from=builder /bin/server /server
COPY --from=builder --chown=65532:65532 /etc/authorized_keys /etc/authorized_keys
COPY --from=builder --chown=65532:65532 /data/ssh /data/ssh

# Run as non-root
USER 65532:65532

ENV HOST="0.0.0.0"
ENV PORT="2222"
ENV AUTHORIZED_KEYS_PATH="/etc/authorized_keys"
ENV SSH_HOST_KEY_PATH="/data/ssh/id_ed25519"

EXPOSE 2222
ENTRYPOINT ["/server"]
