# syntax=docker/dockerfile:1

## Build
FROM golang:1.17-alpine as build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download \
  && mkdir /app/assets

COPY *.go ./
COPY assets/* /app/assets/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /app/cryptopricey

## Deploy
# hadolint ignore=DL3006
FROM alpine:3

RUN apk add --no-cache musl-dev go

# Configure Go
ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH

RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin \
  && mkdir -p /app/assets \
  && chmod -R 0755 /app

WORKDIR /app

COPY --from=build /app/cryptopricey /app/cryptopricey

RUN addgroup -g 1000 -S cryptopricey && \
    adduser -u 1000 -S cryptopricey -G cryptopricey

RUN chmod 0755 /app/cryptopricey

USER cryptopricey

EXPOSE 8080

CMD ["/app/cryptopricey"]
