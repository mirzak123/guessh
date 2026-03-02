# Stage 1: Build
FROM alpine:3.19 AS builder
RUN apk add --no-cache gcc git make musl-dev cjson-dev pkgconfig

WORKDIR /app
COPY . .
RUN make build-server

# Stage 2: Runtime
FROM alpine:3.19
RUN apk add --no-cache cjson
WORKDIR /app
COPY --from=builder /app/bin/server .
COPY words/ ./words/

ENV GUESSH_SERVER_PORT=2480
ENV WORDS_PATH=./words

EXPOSE 2480
CMD ["./server"]
