# Use official Go image to build, then a minimal image to run
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /pr-guardian-action ./cmd/action

# Final minimal image
FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /pr-guardian-action /pr-guardian-action
ENTRYPOINT ["/pr-guardian-action"]
