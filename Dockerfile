FROM golang:1.21-alpine as build

WORKDIR /everest

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /everest-api cmd/main.go

FROM scratch

WORKDIR /

COPY --from=build /everest-api /everest-api
COPY migrations /migrations

EXPOSE 8080

ENTRYPOINT ["/everest-api"]
