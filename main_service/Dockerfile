FROM golang:latest

COPY ./ /app


RUN apt install openssl

WORKDIR /app/main_service

RUN openssl genrsa -des3 -out /app/main_service/encrypted.pem -passout pass:"someshit" 2048
RUN openssl rsa -in /app/main_service/encrypted.pem  -outform PEM -pubout -out /app/main_service/signature.pub -passin pass:"someshit"
RUN openssl rsa -in /app/main_service/encrypted.pem -out /app/main_service/signature.pem -outform PEM -passin pass:"someshit"

RUN go mod download -x

RUN go build main.go

ENTRYPOINT ./main