FROM golang:alpine
ENV LANGUAGE="en"
COPY . /build
WORKDIR /build
RUN go build -o mytelegrambot
EXPOSE 80/tcp
CMD [ "./code" ]