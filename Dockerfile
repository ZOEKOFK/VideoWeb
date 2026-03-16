FROM golang:1.24-alpine AS builder
WORKDIR /app
# 设置国内 Go 代理
ENV GOPROXY=https://goproxy.cn,direct
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/main .
RUN mkdir -p uploads/videos uploads/avatars
EXPOSE 8888
CMD ["./main"]