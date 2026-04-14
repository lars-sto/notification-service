FROM golang:1.25.4-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/notification-service ./cmd/notification-service

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /
COPY --from=builder /out/notification-service /notification-service

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/notification-service"]