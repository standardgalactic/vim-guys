# Use official Golang image to build the app
FROM golang:1.24-alpine as builder

WORKDIR /app
COPY . .

# Build the Go app
RUN go mod download
RUN go build -o main

# Create a minimal image for deployment
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080

CMD ["./main"]

