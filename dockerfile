FROM golang:alpine
ENV LANGUAGE="en"
COPY /code/code .
WORKDIR /code
RUN go build -o mytelegrambot
EXPOSE 80/tcp
CMD [ "./code" ]