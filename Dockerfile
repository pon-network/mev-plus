# Stage 1: Building the application
# Using the official Golang 1.20 image as the base image for building the app
FROM golang:1.22.2 AS builder

# the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies.
# If the go.mod and the go.sum file are not changed, then the docker build cache
# will not re-run this step, thereby saving time
RUN go mod download

# Copy the entire MEV Plus project to the container
COPY . .

# Compile the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mevPlus ./mevPlus.go

# Stage 2: Setup the runtime container
# Use a minimal base image to reduce the final image size and attack surface
FROM alpine:latest

# Set the working directory in the container
WORKDIR /root/

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/mevPlus .
COPY --from=builder /app/entrypoint.sh .
COPY --from=builder /app/setup-wizard.yml .
COPY --from=builder /app/avatar-default.png .

# Make sure the entrypoint script is executable
RUN chmod +x ./entrypoint.sh

# K2 server port
EXPOSE 10000

ENTRYPOINT [ "./entrypoint.sh" ]