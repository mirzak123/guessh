# Stage 1: Build
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY cli/go.mod cli/go.sum ./
RUN cd . && go mod download
COPY cli/ ./
RUN go build -o /guessh-sshd ./cmd/guessh-sshd/main.go

# Stage 2: Runtime
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /guessh-sshd .
COPY words/ ./words/

EXPOSE 2222
CMD ["./guessh-sshd"]
