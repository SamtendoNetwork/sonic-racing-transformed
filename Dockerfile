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

# * Authentication and secure servers. Both are UDP; publishing them as TCP
# * silently produces a server nothing can reach.
EXPOSE 60000/udp
EXPOSE 60001/udp

# * Leaderboards and common data live here; mount a volume to keep them.
VOLUME /app/data

# * Docker's userland proxy rewrites the source address, which would make the
# * server reflect the proxy's address back to consoles instead of their own
# * and break peer-to-peer connections. Run with --network host, or make sure
# * the published UDP ports preserve the source address.
CMD ["./sonic-nex"]
