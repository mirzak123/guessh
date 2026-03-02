# Stage 1: Build
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY cli/go.mod cli/go.sum ./
RUN cd . && go mod download
COPY cli/ ./
RUN go build -o /guessh-ssh ./cmd/ssh/main.go

# Stage 2: Runtime
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /guessh-ssh .
COPY words/ ./words/

EXPOSE 2222
CMD ["./guessh-ssh"]
