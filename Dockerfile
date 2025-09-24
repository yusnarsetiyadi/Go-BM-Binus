FROM golang:1.24.3-alpine AS build

RUN mkdir /app

WORKDIR /app

COPY ./ /app

RUN go mod tidy

RUN go build -o bm_binus

EXPOSE 80

CMD [ "./bm_binus" ]