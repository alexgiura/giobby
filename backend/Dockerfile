# Build stage
FROM golang:1.22.12-alpine AS builder

WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o migrate ./cmd/migrate

# Final stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/migrate .
COPY swagger.json /root/swagger.json

ENV SWAGGER_PATH=/root/swagger.json

EXPOSE 8080

CMD ["./main"]
