FROM golang:1.24 AS builder
RUN apt-get update

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY main.go ./
RUN go build -o /app/main .
RUN chmod +x /app/main

# FROM alpine:3.14
# RUN mkdir -p /app
# COPY --from=builder /app/main /app/awscost
# # WORKDIR /app
# RUN chmod +x /app/awscost

CMD ["/app/main"]
# CMD ["echo", "$AWS_ACCESS_KEY_ID"]
# # CMD ["ls", "-l", "/app"]
# CMD ["bash", "-c", "/app/awscost"]