FROM golang:1.22.12 AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/fake-gpu-plugin ./cmd/fake-gpu-plugin
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/nvidia-smi ./cmd/fake-nvidia-smi

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /out/fake-gpu-plugin /usr/local/bin/fake-gpu-plugin
COPY --from=builder /out/nvidia-smi /usr/local/bin/nvidia-smi
ENTRYPOINT ["/usr/local/bin/fake-gpu-plugin"]
