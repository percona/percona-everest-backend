FROM golang:1.20-alpine as build

WORKDIR /cmd

COPY . .

RUN go build -o /everest-api .

FROM alpine:latest

WORKDIR /

COPY --from=build /everest-api /everest-api

EXPOSE 8081

USER nonroot:nonroot

ENTRYPOINT ["/everest-api"]
