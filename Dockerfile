# syntax=docker/dockerfile:1
# Multi-stage build: pure Go (CGO off) binary in a distroless, non-root image.

FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown
RUN CGO_ENABLED=0 go build -trimpath \
      -ldflags "-s -w \
        -X github.com/PaulPerigault/rsa-ct-analyzer/internal/version.Version=${VERSION} \
        -X github.com/PaulPerigault/rsa-ct-analyzer/internal/version.Commit=${COMMIT} \
        -X github.com/PaulPerigault/rsa-ct-analyzer/internal/version.Date=${DATE}" \
      -o /out/rsa-ct-analyzer ./cmd/rsa-ct-analyzer \
 && mkdir /out/work && chown 65532:65532 /out/work

FROM gcr.io/distroless/static-debian12:nonroot
LABEL org.opencontainers.image.title="rsa-ct-analyzer" \
      org.opencontainers.image.description="Audit RSA keys from certificates and Certificate Transparency logs" \
      org.opencontainers.image.source="https://github.com/PaulPerigault/rsa-ct-analyzer" \
      org.opencontainers.image.licenses="MIT"
COPY --from=build /out/rsa-ct-analyzer /rsa-ct-analyzer
COPY --from=build --chown=65532:65532 /out/work /work
WORKDIR /work
USER 65532:65532
ENTRYPOINT ["/rsa-ct-analyzer"]
CMD ["--help"]
