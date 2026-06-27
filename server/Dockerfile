# Builder Stage
FROM golang:1.26.3 AS builder

WORKDIR /workspace

COPY go.work .
COPY shared/go.mod ./shared/
COPY server/go.mod ./server/
COPY agent/go.mod ./agent/
COPY dashboard/go.mod ./dashboard/

WORKDIR /workspace/server

RUN go mod download

WORKDIR /workspace

COPY . .

WORKDIR /workspace/server

RUN CGO_ENABLED=0 GOOS=linux \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /server .


# Runtime Stage
FROM scratch

COPY --from=builder /server /server

EXPOSE 8080

ENTRYPOINT ["/server"]