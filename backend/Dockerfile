FROM golang:1.26.1-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o filelogix-backend .

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /app/filelogix-backend /filelogix-backend
COPY --from=builder /app/database/migrations ./database/migrations

USER nonroot:nonroot

CMD ["/filelogix-backend"]