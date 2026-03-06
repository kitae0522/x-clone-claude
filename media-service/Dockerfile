FROM golang:1.24

RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    webp \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

RUN go install github.com/air-verse/air@v1.52.3

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8081

CMD ["air", "-c", ".air.toml"]
