# Contributing

Bug reports, protocol traces with credentials removed, documentation fixes and
tested protocol implementations are welcome.

Before opening a pull request:

1. Run `gofmt -w` on changed Go files.
2. Run `go vet ./...`.
3. Run `go test ./...`.
4. Do not commit account passwords, `.env` files, player data, packet captures
   containing tokens, or Nintendo/SEGA game files.

Please explain the observed Wii U behaviour and the evidence behind wire-format
changes. A small regression test is strongly preferred when practical.
