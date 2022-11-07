FROM golang:latest

WORKDIR /build/metrics

COPY . .

RUN go get -d -v
RUN go build -v

CMD ["./metrics"]