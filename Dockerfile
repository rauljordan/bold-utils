# Use an official Golang runtime as a parent image
FROM golang:1.20

# Set the working directory inside the container
WORKDIR /app

# Copy the local package files to the container's workspace.
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code from the current directory to the /app directory in the container
COPY . .

# Build the application
RUN go build -o main .

# Run the executable
CMD ["./main"]