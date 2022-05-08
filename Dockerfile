# Development image with go toolchain
FROM golang:1.14-alpine AS builder

RUN apk update && \
    apk add alpine-sdk && \
    rm -rf /var/cache/apk/*

RUN mkdir -p /app
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -o bin/restic-controller .

# Actual released image
FROM restic/restic:0.13.1

RUN mkdir -p /app
WORKDIR /app
COPY --from=builder /app/bin/restic-controller .

EXPOSE 8080

ENTRYPOINT ["./restic-controller"]
