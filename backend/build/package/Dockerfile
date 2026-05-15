FROM node:22-alpine AS frontend-builder

ARG BASE_PATH=/client

WORKDIR /src/frontend
COPY frontend/ ./
COPY frontend/package*.json ./
RUN npm ci
RUN BASE_PATH=${BASE_PATH} npm run build

FROM golang:1.26-alpine3.22 AS backend-builder

ARG BUILD_TIME=""
ARG VCS_REF=""
ARG BASE_PATH=/client

WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./

COPY --from=frontend-builder /src/frontend/dist /src/backend/internal/webclient/dist

RUN go run ./cmd/gencert

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

RUN go build \
    -ldflags="-s -w \
    -X github.com/willie68/schematics2/backend/internal/version.BuildTime=${BUILD_TIME} \
    -X github.com/willie68/schematics2/backend/internal/version.Commit=${VCS_REF} \
    -X github.com/willie68/schematics2/backend/internal/version.ClientBasePath=${BASE_PATH}" \
    -o /out/schematics2 ./cmd/server

FROM alpine:3.22

ARG RELEASE
ENV IMG_VERSION="${RELEASE}"

RUN apk add --no-cache ca-certificates wget

WORKDIR /app

COPY --from=backend-builder /out/schematics2 /usr/local/bin/schematics2
COPY --from=backend-builder /src/backend/configs /app/configs

RUN chmod 0755 /usr/local/bin/schematics2

ENTRYPOINT ["/usr/local/bin/schematics2"]

EXPOSE 8080 8443

HEALTHCHECK --interval=30s --timeout=5s --retries=3 --start-period=10s \
  CMD wget -q -T 5 --spider http://localhost:8080/readyz || exit 1

LABEL org.opencontainers.image.title="schematics2" \
      org.opencontainers.image.description="schematics2 backend service" \
      org.opencontainers.image.version="${IMG_VERSION}" \
      org.opencontainers.image.source="https://github.com/willie68/schematics2" \
      org.opencontainers.image.vendor="willie68" \
      NAME="schematics2"

