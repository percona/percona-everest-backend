FROM golang:1.20-alpine

WORKDIR /cmd

COPY . .

RUN go build -o everest-api .

CMD ["./everest-api"]
