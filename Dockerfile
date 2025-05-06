FROM golang:1.24 AS builder
RUN apt-get update

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY main.go ./
RUN go build -o /app/main .
RUN chmod +x /app/main

FROM alpine:3.14
RUN apk add libc6-compat
RUN mkdir -p /app
COPY --from=builder /app/main /app/awscost
# # WORKDIR /app
RUN chmod +x /app/awscost

# ENTRYPOINT ["/app/main"]
ENTRYPOINT ["/app/awscost"]