FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

FROM alpine:latest

RUN apk add --no-cache \
    ffmpeg \
    imagemagick \
    ca-certificates \
    tzdata

WORKDIR /root/

COPY --from=builder /app/main .

RUN mkdir -p /tmp/compression

EXPOSE 3000

CMD ["./main"]
