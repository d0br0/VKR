FROM golang:alpine
ENV LANGUAGE="en"
COPY . /app
WORKDIR /app
RUN go build -o mytelegrambot
EXPOSE 80/tcp
CMD [ "./code" ]