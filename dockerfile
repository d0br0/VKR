FROM alpine:3.19
ENV LANGUAGE="en"
COPY /code/code .
RUN apk add --no-cache ca-certificates &&\
    chmod +x code
EXPOSE 80/tcp
CMD [ "./code" ]