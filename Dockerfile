# Build environment
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache ca-certificates

WORKDIR /build

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# Build with optimizations and static linking
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o go-irr \
    .

# Runtime environment
FROM alpine:edge AS runtime

# Add testing repository (for bgpq4)
RUN echo "@testing http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories
RUN apk add --no-cache bgpq4@testing

# Run as a non-root user
RUN addgroup -S irr && adduser -S irr -G irr

USER irr

WORKDIR /home/irr

COPY --from=builder --chown=irr:irr /build/go-irr ./go-irr

RUN chmod +x ./go-irr

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./go-irr"]

EXPOSE 8080
