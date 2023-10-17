FROM golang:1.21-alpine as build

WORKDIR /everest

COPY . .

RUN apk add -U --no-cache ca-certificates

FROM scratch

WORKDIR /

COPY ./bin/percona-everest-backend  /everest-api
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY migrations /migrations

EXPOSE 8080

ENTRYPOINT ["/everest-api"]
