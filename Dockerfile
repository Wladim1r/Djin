# Build step
FROM golang:1.24.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy && go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/cmd/djin ./cmd/main.go

# Final step
FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/cmd/djin .

COPY --from=builder /app/web ./web

COPY --from=builder /app/internal/db ./internal/db

EXPOSE 8080

CMD ["./djin"]
