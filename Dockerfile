FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o mockspotify-server ./cmd/server
RUN go build -o mockspotify-seed ./cmd/seed

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/mockspotify-server .
COPY --from=builder /app/mockspotify-seed .
COPY templates/ templates/
COPY static/ static/
RUN mkdir -p data
EXPOSE 8090
CMD ["./mockspotify-server"]
