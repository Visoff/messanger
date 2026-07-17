FROM rust:1.85-alpine AS builder
WORKDIR /app
RUN apk add --no-cache musl-dev pkgconfig openssl-dev
COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo "fn main() {}" > src/main.rs && cargo build --release 2>/dev/null || true
COPY src/ src/
RUN cargo build --release

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/target/release/file_storage /app/file_storage
EXPOSE 3001
CMD ["/app/file_storage"]
