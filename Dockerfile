# --- STAGE 1: Builder ---
# Use the official Golang Alpine image as the base for building.
# Alpine is a lightweight Linux distribution.
FROM golang:alpine AS builder

# Set the working directory inside the container.
WORKDIR /app

# Copy go.mod and go.sum files first.
# This layers the download step and leverages Docker cache to speed up future builds.
COPY go.mod go.sum ./

# Download all dependencies.
RUN go mod download

# Copy the rest of the source code into the container.
COPY . .

# Build the Go app as a statically linked binary.
# CGO_ENABLED=0 ensures the binary can run on a minimal OS without C libraries.
# GOOS=linux sets the target operating system.
RUN CGO_ENABLED=0 GOOS=linux go build -o core_engine_api main.go


# --- STAGE 2: Runner ---
# Use a minimal Alpine image for the final production environment.
FROM alpine:latest

# Install timezone data and CA certificates (good practice for secure API calls and time handling).
RUN apk --no-cache add ca-certificates tzdata

# Set the working directory.
WORKDIR /root/

# Copy only the compiled binary from the builder stage.
# This leaves all source code and heavy Go tools behind, keeping the image tiny.
COPY --from=builder /app/core_engine_api .

# Expose port 8080 to the outside world.
EXPOSE 8080

# Command to run the executable.
CMD ["./core_engine_api"]