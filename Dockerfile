# Use the official Golang image as the base image
FROM golang:1.21.6-alpine

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o local-content-server .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./local-content-server"]