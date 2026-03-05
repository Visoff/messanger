from golang:1.26 as builder

workdir /app

copy . .

run go mod download

run go build -o main cmd/main/main.go

from debian:stable-slim

copy --from=builder /app/main /app/main

cmd ["/app/main"]
