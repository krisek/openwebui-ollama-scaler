# Build stage
FROM golang:1.23 as builder

# Set the Current Working Directory inside the container
RUN mkdir /server 
WORKDIR /server

# Copy the source code into the container
COPY openwebui-ollama-scaler.go /server

# Initialize the Go module and download dependencies
# RUN go mod init gopkg.in/yaml.v2 \
#    && go mod tidy
ENV GO111MODULE=on
ENV GOROOT=/usr/local/go
RUN go mod init server && go mod tidy

# Build the Go app
RUN CGO_ENABLED=0 go build -o openwebui-ollama-scaler openwebui-ollama-scaler.go

# Final stage
FROM gcr.io/distroless/base

# Copy the Go binary from the builder stage
COPY --from=builder /server/openwebui-ollama-scaler /openwebui-ollama-scaler

# Use an unprivileged user
USER 65534:65534

# Command to run the executable
ENTRYPOINT ["/openwebui-ollama-scaler"]
