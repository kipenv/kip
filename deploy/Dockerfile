FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o /kip-server ./server/cmd/server/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /kip-server /usr/local/bin/kip-server

EXPOSE 8080
ENTRYPOINT ["kip-server"]
