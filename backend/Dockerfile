FROM golang:1.24-alpine

WORKDIR /app

RUN go install github.com/air-verse/air@v1.52.3

COPY . .

RUN go mod download

CMD ["air", "-c", ".air.toml"]