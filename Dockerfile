FROM golang:1.20-alpine as build

WORKDIR /everest

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /everest-api cmd/main.go

FROM alpine:latest

WORKDIR /

COPY --from=build /everest-api /everest-api
COPY migrations /migrations

EXPOSE 8081

ENTRYPOINT ["/everest-api"]
