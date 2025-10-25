# syntax=docker/dockerfile:1

FROM golang:1.25.1-bookworm AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /app/server /app/server
EXPOSE 8080
CMD ["./server"]
