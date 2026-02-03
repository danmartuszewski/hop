FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Run tests
FROM builder AS tester
RUN go test -v ./...

# Build binary
FROM builder AS build
RUN go build -o /hop ./cmd/hop

# Final image
FROM alpine:3.19

RUN apk add --no-cache \
    openssh-client \
    bash

COPY --from=build /hop /usr/local/bin/hop

# Create a non-root user
RUN adduser -D -s /bin/bash hopuser

# Add example config
RUN mkdir -p /home/hopuser/.config/hop
COPY example-config.yaml /home/hopuser/.config/hop/config.yaml
RUN chown -R hopuser:hopuser /home/hopuser/.config

USER hopuser
WORKDIR /home/hopuser

CMD ["/bin/bash"]
