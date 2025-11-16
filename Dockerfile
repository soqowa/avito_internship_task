FROM golang:1.22-alpine AS builder

ENV GOTOOLCHAIN=auto

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/reviewer-svc ./cmd/reviewer-svc

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /bin/reviewer-svc /app/reviewer-svc
COPY --from=builder /app/migrations /app/migrations

ENV PORT=8080

EXPOSE 8080

ENTRYPOINT ["/app/reviewer-svc"]
