FROM golang:1.17

# Install necessary tools
RUN apt-get update && apt-get install -y \
    git \
    && rm -rf /var/lib/apt/lists/*

# Set the working directory
WORKDIR /workspace

# Copy go.mod and go.sum to the workspace
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Install the application
RUN go install ./...

# Expose the application port (if needed)
EXPOSE 8080
