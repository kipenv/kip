FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Stamped into server/cmd/server/main.go (`var version`). Optional:
#   docker build --build-arg VERSION=0.1.0 -f deploy/Dockerfile .
ARG VERSION=dev
RUN CGO_ENABLED=1 go build -ldflags "-s -w -X main.version=${VERSION}" -o /kip-server ./server/cmd/server/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates \
    && adduser -D -u 1000 kip \
    && mkdir /data && chown kip:kip /data
COPY --from=builder /kip-server /usr/local/bin/kip-server

USER kip
EXPOSE 8080
ENTRYPOINT ["kip-server"]
