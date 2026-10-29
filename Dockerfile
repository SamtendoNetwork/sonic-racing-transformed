FROM golang:1.24-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
COPY third_party ./third_party
RUN go mod download

COPY . .
RUN go build -o /out/sonic-nex .

FROM alpine:3.20

WORKDIR /app

COPY --from=build /out/sonic-nex .

EXPOSE 60000/udp
EXPOSE 60001/udp

VOLUME /app/data

CMD ["./sonic-nex"]
