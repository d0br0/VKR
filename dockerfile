FROM golang:alpine
ENV LANGUAGE="en"
COPY . /app
WORKDIR /app
RUN go build -o mytelegrambot
EXPOSE 8080/tcp
CMD [ "./code" ]