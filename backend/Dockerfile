FROM node:22-alpine AS frontend-builder
WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -o main cmd/main/main.go

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/main /app/main
COPY --from=frontend-builder /app/build /app/static
ENV STATIC_DIR=/app/static
CMD ["/app/main"]
