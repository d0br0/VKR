FROM golang:alpine
ENV LANGUAGE="en"
COPY . .
WORKDIR /app
RUN go build -o mytelegrambot
EXPOSE 80/tcp
CMD [ "./code" ]