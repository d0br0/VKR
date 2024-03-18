FROM golang:alpine
ENV LANGUAGE="en"
COPY . .
WORKDIR /app
RUN go build -o mybot
CMD [ "./code" ]