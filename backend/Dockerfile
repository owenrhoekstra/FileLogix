FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o filelogix-backend .

FROM alpine:3.21

RUN apk add --no-cache postgresql16-client rclone

RUN addgroup -g 65532 nonroot && adduser -u 65532 -G nonroot -s /sbin/nologin -D nonroot

WORKDIR /app
COPY --from=builder /app/filelogix-backend /filelogix-backend

USER nonroot:nonroot
CMD ["/filelogix-backend"]