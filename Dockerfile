FROM golang:1.13-alpine AS builder
RUN apk add --no-cache ca-certificates git

RUN mkdir /build
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -o miniflux-substack-filter ./cmd

FROM alpine:3.11.3
COPY --from=builder /build/miniflux-substack-filter .

ENTRYPOINT [ "./miniflux-substack-filter" ]
