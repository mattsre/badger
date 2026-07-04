# syntax=docker/dockerfile:1

FROM cgr.dev/chainguard/go:latest AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /badger .

FROM cgr.dev/chainguard/static:latest

COPY --from=builder /badger /badger

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/badger"]
