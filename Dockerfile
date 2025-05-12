FROM golang:latest AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN git rev-parse --short HEAD > /tmp/git_hash || echo "unknown" > /tmp/git_hash

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-X main.GitCommit=$(cat /tmp/git_hash)" -o creemproxy .

FROM alpine:latest
WORKDIR /app/
COPY --from=builder /app/creemproxy .
COPY --from=builder /tmp/git_hash /app/git_hash

ENTRYPOINT ["sh", "-c", "cat /app/git_hash && /app/creemproxy"]

EXPOSE 8443