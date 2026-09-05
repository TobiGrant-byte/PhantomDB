FROM golang:1.26 AS builder
WORKDIR /build
COPY . .
WORKDIR /build/phantomdb-server
RUN CGO_ENABLED=0 GOOS=linux go build -o /phantomdb-server .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /phantomdb-server .
EXPOSE 8080
VOLUME /app/data
CMD ["./phantomdb-server", "--db", "/app/data/phantomdb-data", "--port", "8080"]