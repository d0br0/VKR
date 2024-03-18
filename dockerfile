FROM golang:alpine
ENV LANGUAGE="en"
COPY . /app
WORKDIR /app
RUN apk add --no-cache ca-certificates &&\
    chmod +x code
EXPOSE 80/tcp
CMD [ "./code" ]