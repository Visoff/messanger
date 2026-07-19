FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 go build -o main .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/main /app/main
EXPOSE 8081 3478/udp
CMD ["/app/main"]
