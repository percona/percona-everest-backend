FROM golang:1.20-alpine as build

WORKDIR /everest

COPY . .
RUN apk update && apk add git

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /everest-api cmd/main.go
RUN git config --global url."https://percona-platform-robot:${ROBOT_TOKEN}@github.com".insteadOf "https://github.com"

FROM scratch

WORKDIR /

COPY --from=build /everest-api /everest-api
COPY migrations /migrations

EXPOSE 8081

ENTRYPOINT ["/everest-api"]
