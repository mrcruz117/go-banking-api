# Use the official Golang image as a parent image
FROM golang:1.22-alpine

# Set the working directory inside the container
WORKDIR /app

COPY . .

# Download all dependencies
RUN go get -d -v ./...

COPY go.mod go.sum ./

# Build the application
RUN go build -o main .

# Expose port 8088
EXPOSE 8088

# Run the application
CMD ["./main"]