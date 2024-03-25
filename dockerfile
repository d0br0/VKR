FROM golang:latest
WORKDIR /app
COPY . .
RUN go build -o mytelegrambot
CMD ["./mytelegrambot"]