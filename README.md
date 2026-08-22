# Sonic & All-Stars Racing Transformed — NEX server

A standalone NEX game server for the Wii U version of *Sonic & All-Stars Racing
Transformed*, written in Go against [Pretendo Network](https://github.com/PretendoNetwork)'s
NEX libraries.

It runs on its own: no PostgreSQL, no gRPC account service, no external
dependencies of any kind beyond the Go toolchain.

```bash
go run .
```

## The game's server identity

These are the values that make the server *this game's* server rather than any
other, taken from the NEX game server list at
[kinnay.github.io](https://kinnay.github.io/view.html?page=nexwiiu) (listed
there as "Sonic Transformed"):

| | |
|---|---|
| Game server ID | `0x10111F00` |
| Access key | `b26a3421` |
| Title IDs | `000500001010B300` (USA), `0005000010111F00` (EUR), `000500001015B400` (JPN) |
| Server build string | `branch:origin/release/ngs/3.4.x.3 build:3_4_13_3_0` |
| Original server | `34.208.166.202:43240` (offline since April 2024) |

The **access key** matters most: every PRUDP packet carries a signature derived
from it, so a console using a different key is ignored at the transport layer.

### NEX version: 3.0.1, not 3.4.13

The build string above says `3.4.13.3`, and it is tempting to configure the
server as NEX 3.4.13. That would be wrong, and it would corrupt matchmaking.

`3.4.13.3` is the version of the *server* SEGA rented from Nintendo — around
nineteen unrelated third-party Wii U titles all reported the same one. The
version that determines the wire format is the NEX *library linked into the
game*, and kinnay's index of Wii U executables
([wiiu.json](https://kinnay.github.io/data/wiiu.json)) shows every Sonic
Transformed build linking:

```
nex 3.0.1     core: Secure Connection, NAT Traversal
nexmm 3.0.1   matchmaking
nexrk 3.0.1   ranking
```

Notably **absent**: `nexds` (DataStore) and `nexut` (Utility). The game cannot
call those protocols, so this server does not implement them — and registering
Utility would be actively harmful, since three different protocols claim ID
`0x6E`.

The practical consequences of 3.0.1 versus 3.4.13 are exact: at 3.0.1 a
`MatchmakeSession` has no `ProgressScore` field, `MatchmakeSessionSearchCriteria`
has no `VacantParticipants`, `NotificationEvent` stops after `StrParam`, and
`RVConnectionData` carries no timestamp. Configure 3.4.13 and every one of those
shifts by a few bytes.

## Protocols

| ID | Protocol | Coverage |
|---|---|---|
| `0x0A` | Ticket Granting | `LoginEx`, `RequestTicket`, `GetPID`, `GetName`. Legacy `Login` is refused by design — the Wii U uses `LoginEx`. |
| `0x0B` | Secure Connection | All eight methods. |
| `0x03` | NAT Traversal | All seven methods. |
| `0x15` | Match Making | 22 methods: the full gathering lifecycle, participants, peer URLs, state, host migration. |
| `0x32` | Match Making Ext | 5 of 6 methods, including `GetParticipantsURLs`. |
| `0x6D` | Matchmake Extension | 22 methods: create, join, browse, auto-matchmake, session mutation, notification data. |
| `0x70` | Ranking | 11 methods: score upload, all five ranking modes, common data, approximate order. |
| `0x0E` | Notifications | Server→client push, used for participation and ownership events. |

All three matchmaking protocols are registered because all three live inside
`nexmm` and nothing in the executable reveals which the game calls. They share
one store, so either route produces the same lobbies.

Communities (persistent gatherings) and block lists are not implemented; if the
game turns out to use them, they surface in the log as `Core::NotImplemented`
with the exact protocol and method.

### How a race actually happens

The server never carries gameplay traffic. It brokers the lobby and then gets
out of the way:

1. Each console publishes its addresses with `SecureConnection::RegisterEx`.
   The server **replaces the console's claimed public address with the address
   the packet actually arrived from**, which is what makes this work behind NAT.
2. A host creates a session; others find it by browsing, by gathering ID, or
   through automatic matchmaking, and join.
3. Joining players call `MatchMakingExt::GetParticipantsURLs` and receive every
   participant's reflected address.
4. `NATTraversal` relays probe requests so the consoles can punch through each
   other's NAT.
5. The race runs peer-to-peer, encrypted with the 32-byte session key the server
   generated and handed to everyone who joined.

## Running it

```bash
cp .env.example .env      # then edit SONIC_SECURE_SERVER_HOST
go run .
```

For an internet-facing deployment, also set a strong
`SONIC_KERBEROS_PASSWORD`, configure either `SONIC_ACCOUNTS_FILE` or a shared
`SONIC_ACCOUNT_SECRET`, and keep real credentials outside the repository. The
example configuration rejects unknown accounts until that integration is
deliberately enabled.

Two UDP ports open: `60000` for authentication, `60001` for the game.

`SONIC_SECURE_SERVER_HOST` must be an address **the console can reach**. It is
sent to the console inside the login response, so `127.0.0.1` only works when
the client is on this machine. Set it to your LAN IP for console testing. Get it
wrong and login succeeds, then nothing happens — the classic symptom.

See [.env.example](.env.example) for every setting.

## Getting a console to connect

This is the part the server cannot do for you, and it is worth being blunt about
it: **a Wii U will not talk to this server until something tells it to.**

The console learns which game server to use, and the credentials to use, from an
account server:

```
GET https://account.nintendo.net/v1/api/provider/nex_token/@me?game_server_id=10111F00
→ host, port, pid, nex_password, token
```

Nintendo's account servers shut down in April 2024, and Pretendo does not host
this title. So you need either a redirected account server that returns this
server's address, or a client that lets you point it directly — which is what
[tests/e2e_test.py](tests/e2e_test.py) does.

The password is the part that must line up. The console never derives it; it is
*issued* by the account server, and this server has to know the same string.
Two supported ways to arrange that:

- **`SONIC_ACCOUNTS_FILE`** — a JSON map of PID to password, for when the
  passwords are already fixed.
- **`SONIC_ACCOUNT_SECRET`** — derivation. Any PID gets
  `HMAC-SHA256(secret, little_endian_uint64(pid))` encoded as unpadded Base64URL,
  so an account server sharing the secret computes the same password with no
  shared database.

If the passwords disagree there is **no server-side error**: the console cannot
decrypt the Kerberos ticket and simply stops. Check this first when a real
console goes quiet after login.

## The structure header problem

The one wire-format decision that cannot be looked up, and the reason this
server detects rather than assumes.

NEX structures may or may not be prefixed with a version byte and a length. The
usual rule — headers arrive in NEX 3.5.0 — would put this game (3.0.1) firmly on
the "no headers" side. But that rule describes when the *feature* appeared, not
when it is *used*, and the two sources disagree for this title: kinnay's own
default profile for Wii U game servers uses PRUDP minor version 4, which turns
headers on.

Reading the code settles the mechanism. The console decides, unilaterally, via
the PRUDP minor version it announces at connection time:

- kinnay's client enables headers when the negotiated minor version is ≥ 3
  (`nintendo/nex/rmc.py`)
- nex-go echoes back whatever minor version the client asked for, with the
  comment *"No change needed, we can just support what the client wants"*

Since no packet capture of this game has ever been published, guessing has a
50% failure rate — and failure means every structure is misparsed.

So the server works it out instead. The login data the console sends is wrapped
in a `DataHolder`, whose framing is identical either way, but whose contents are
a structure. Walking the header chain and checking whether it accounts for the
object exactly is unambiguous. The decision is applied to both servers and
logged:

```
Structure headers enabled (detected from the console's login data)
```

Override with `SONIC_STRUCTURE_HEADER=on|off` if you ever need to pin it.

## Testing

The test suite drives the server with
[kinnay's NintendoClients](https://github.com/kinnay/NintendoClients) — the same
library used to talk to the real Nintendo servers — so it exercises the actual
PRUDP/RMC wire protocol rather than a mock.

```bash
go test ./...                    # unit tests, including the header detector

python -m venv .venv             # Python 3.12+
.venv/Scripts/pip install pycryptodome anyio pyopenssl multidict
.venv/Scripts/pip install --no-deps anynet NintendoClients
cp tests/netifaces.py .venv/Lib/site-packages/

go run . &
.venv/Scripts/python tests/e2e_test.py
```

34 checks covering Kerberos login, the handoff to the secure server, station URL
reflection, session create/browse/join, participant and peer URL lookup, session
mutation, host migration, leaderboards, leaving, gathering reaping on
disconnect, and automatic matchmaking.

`tests/netifaces.py` is a stub: `anynet` imports `netifaces` only for local
interface discovery, and the real package has no wheel for current Python on
Windows.

The suite runs against **both** wire formats — `E2E_MINOR_VERSION=4` for
structure headers and `E2E_MINOR_VERSION=2` for without — and passes all 34
checks either way, with the server detecting correctly each time. That is the
evidence behind the claim that the header ambiguity above is actually handled
rather than merely guessed at.

## Design notes

**Why no database.** Pretendo backs matchmaking with PostgreSQL, which is right
for a service hosting thousands of concurrent sessions. A gathering is inherently
ephemeral — every session dies when the server restarts regardless of where it
was stored — so this server keeps them in a map, modelled directly on common-go's
semantics including the participation notification events. Leaderboards are the
one thing worth keeping, and they persist to `data/ranking.json` via
write-then-rename.

**What is reused.** Ticket granting, secure connection, NAT traversal and ranking
come from `nex-protocols-common-go`, which needs no database for any of them.
Matchmaking is hand-written against `nex-protocols-go` because common-go's
implementation is SQL-bound.

**Friends.** Wii U friend lists live on a separate system-wide friends server, so
`globals.GetUserFriendPIDs` reports that nobody has friends. Friends-only
gatherings therefore cannot be joined by strangers and friend-scoped
leaderboards come back empty — failing closed rather than leaking. Both stores
take the lookup as a function field, so wiring in a real friends server is a
one-line change.

## Layout

```
main.go                        startup, configuration validation
globals/                       config, accounts, friends stub
nexserver/                     the two servers and protocol registration
  authentication.go            ticket granting server
  secure.go                    game server
  protocols_*.go               protocol wiring
  handlers_*.go                hand-written methods
  structure_header.go          wire-format detection
matchmaking/                   in-memory gatherings + the three MM protocols
ranking/                       persistent leaderboards and common data
tests/e2e_test.py              end-to-end suite driven by NintendoClients
```

## Credits

Built on [Pretendo Network](https://github.com/PretendoNetwork)'s `nex-go`,
`nex-protocols-go` and `nex-protocols-common-go`, and on the protocol
documentation and client in
[kinnay/NintendoClients](https://github.com/kinnay/NintendoClients). This is a
preservation project for a game whose servers no longer exist; it is not
affiliated with Nintendo or SEGA.

## License

This project is licensed under the GNU Affero General Public License, version 3
or later. See [LICENSE](LICENSE) and [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).
