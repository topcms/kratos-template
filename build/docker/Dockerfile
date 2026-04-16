# syntax=docker/dockerfile:1
FROM golang:1.26-bookworm AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/server ./server
COPY configs ./configs
USER nonroot:nonroot
EXPOSE 8000 9000
ENTRYPOINT ["./server"]
CMD ["-conf", "./configs"]
