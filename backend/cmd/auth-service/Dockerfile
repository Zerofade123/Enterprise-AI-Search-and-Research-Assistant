FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /auth-service ./cmd/auth-service

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /auth-service /auth-service
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/auth-service"]
