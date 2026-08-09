# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.26.5

FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=development
ARG COMMIT=unknown
ARG CONTRACT_VERSION=1.0.0

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -trimpath \
    -ldflags="-s -w \
      -X github.com/tgwtea/penny-backend/internal/platform/buildinfo.version=${VERSION} \
      -X github.com/tgwtea/penny-backend/internal/platform/buildinfo.commit=${COMMIT} \
      -X github.com/tgwtea/penny-backend/internal/platform/buildinfo.contractVersion=${CONTRACT_VERSION}" \
    -o /out/api ./cmd/api

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -trimpath \
    -ldflags="-s -w \
      -X github.com/tgwtea/penny-backend/internal/platform/buildinfo.version=${VERSION} \
      -X github.com/tgwtea/penny-backend/internal/platform/buildinfo.commit=${COMMIT} \
      -X github.com/tgwtea/penny-backend/internal/platform/buildinfo.contractVersion=${CONTRACT_VERSION}" \
    -o /out/worker ./cmd/worker

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -trimpath \
    -ldflags="-s -w \
      -X github.com/tgwtea/penny-backend/internal/platform/buildinfo.version=${VERSION} \
      -X github.com/tgwtea/penny-backend/internal/platform/buildinfo.commit=${COMMIT} \
      -X github.com/tgwtea/penny-backend/internal/platform/buildinfo.contractVersion=${CONTRACT_VERSION}" \
    -o /out/migrate ./cmd/migrate

FROM scratch AS api
ARG VERSION=development
ARG COMMIT=unknown
LABEL org.opencontainers.image.title="Penny API" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.source="https://github.com/tgwtea/penny-backend"
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/api /penny-api
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/penny-api"]

FROM scratch AS worker
ARG VERSION=development
ARG COMMIT=unknown
LABEL org.opencontainers.image.title="Penny Worker" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.source="https://github.com/tgwtea/penny-backend"
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/worker /penny-worker
USER 65532:65532
ENTRYPOINT ["/penny-worker"]

FROM scratch AS migrate
ARG VERSION=development
ARG COMMIT=unknown
LABEL org.opencontainers.image.title="Penny Database Migrator" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.source="https://github.com/tgwtea/penny-backend"
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/migrate /penny-migrate
COPY db/migrations /migrations
ENV PENNY_MIGRATIONS_SOURCE=file:///migrations
USER 65532:65532
ENTRYPOINT ["/penny-migrate"]
