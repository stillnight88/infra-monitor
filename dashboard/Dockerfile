# Builder Stage
FROM golang:1.26.3 AS builder

WORKDIR /workspace

COPY go.work .
COPY shared/go.mod ./shared/
COPY server/go.mod server/go.sum ./server/
COPY agent/go.mod agent/go.sum ./agent/
COPY dashboard/go.mod dashboard/go.sum ./dashboard/

WORKDIR /workspace/dashboard

RUN go mod download

WORKDIR /workspace

COPY . .

WORKDIR /workspace/dashboard

RUN CGO_ENABLED=0 GOOS=linux \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /dashboard .

# Runtime Stage    
FROM scratch

COPY --from=builder /dashboard /dashboard

ENTRYPOINT ["/dashboard"]