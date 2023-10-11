FROM golang:1.21-alpine as build

WORKDIR /everest

COPY . .

ARG TELEMETRY_URL="https://check.percona.com"
ARG TELEMETRY_INTERVAL="24h"

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -ldflags " \
    -X 'github.com/percona/percona-everest-backend/cmd/config.TelemetryURL=$TELEMETRY_URL'  \
    -X 'github.com/percona/percona-everest-backend/cmd/config.TelemetryInterval=$TELEMETRY_INTERVAL'" \
    -o /everest-api cmd/main.go
RUN apk add -U --no-cache ca-certificates

FROM scratch

WORKDIR /

COPY --from=build /everest-api /everest-api
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY migrations /migrations

EXPOSE 8080

ENTRYPOINT ["/everest-api"]
