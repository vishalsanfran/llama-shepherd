FROM golang:1.24 as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o router ./cmd/router

FROM gcr.io/distroless/base-debian12
COPY --from=builder /app/router /router
ENTRYPOINT ["/router"]