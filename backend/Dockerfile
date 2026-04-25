FROM golang:1.26.2-alpine3.23 AS builder

WORKDIR /src
COPY . .
RUN go mod download
RUN go build -o schematic2-api ./cmd/api

FROM alpine:3.23
WORKDIR /app
COPY --from=builder /src/schematic2-api /usr/local/bin/schematic2-api
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/schematic2-api"]
