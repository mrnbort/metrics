FROM golang:1.19-alpine as build

ADD . /build
WORKDIR /build

RUN  go build -o /build/metrics -ldflags "-s -w"


FROM alpine:3.16

COPY --from=build /build/metrics /srv/metrics
RUN chmod +x /srv/metrics

WORKDIR /srv
EXPOSE 8080
CMD ["/srv/metrics"]
