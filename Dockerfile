# 第一阶段：构建
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o gaokao-server main.go

# 第二阶段：部署
FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/gaokao-server .
RUN chmod +x gaokao-server
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s CMD pgrep gaokao-server || exit 1
ENTRYPOINT ["./gaokao-server"]