# ---------- BUILD STAGE ----------
FROM golang:1.25-alpine AS build

WORKDIR /go/src/logengine
COPY . .

# Build the logengine binary
RUN CGO_ENABLED=0 go build -o /go/bin/logengine ./cmd/logengine

# Download grpc_health_probe
RUN GRPC_HEALTH_PROBE_VERSION=v0.3.2 && \
    wget -q -O /go/bin/grpc_health_probe \
    https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /go/bin/grpc_health_probe

# ---------- RUNTIME STAGE ----------
FROM alpine:3.21 AS runtime

# Install CA certificates (needed if TLS used)
RUN apk add --no-cache ca-certificates

WORKDIR /logengine

# Copy binaries
COPY --from=build /go/bin/logengine /bin/logengine
COPY --from=build /go/bin/grpc_health_probe /bin/grpc_health_probe

# Ensure the data directory exists
RUN mkdir -p /var/run/logengine/data

# Make it a mount point for the PVC
VOLUME /var/run/logengine/data

# Set entrypoint
ENTRYPOINT ["/bin/logengine"]
