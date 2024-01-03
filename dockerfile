# Build Stage
# Use the official Go image from the DockerHub
FROM golang:1.19-alpine as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o indexer cmd/indexer/main.go

# Final Stage
FROM alpine

# expose port 80, 443 to the outside world
EXPOSE 80
EXPOSE 443

# Ensure you copy the binary with the correct name
COPY --from=builder /app/indexer /app/indexer
CMD ["/app/indexer"]
