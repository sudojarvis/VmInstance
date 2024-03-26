# Start with a base image containing the Go runtime
FROM golang:1.22.1-bullseye AS builder


# Set the Current Working Directory inside the container
WORKDIR /app


# Copy the source code from the current directory to the Working Directory inside the container
COPY . .


# Ensure all dependencies are synchronized
RUN go mod tidy


# Download the missing dependency
RUN go mod download google.golang.org/genproto/googleapis/bytestream




# Build the Go app
RUN go build -mod=mod -o /app/main .




# Expose port 8080 to the outside world
EXPOSE 3000


# Command to run the executable
CMD ["/app/main"]
