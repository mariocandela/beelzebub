FROM golang:alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

RUN apk add git

WORKDIR /build

# Download dependency
COPY . .
RUN go mod download


# Build
RUN go build -o main .

WORKDIR /dist

RUN cp /build/main .

# Use scratch image as finally tiny container 
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /dist/main /

ENTRYPOINT ["/main"]