# Third-party notices

This repository vendors modified copies of:

- `github.com/PretendoNetwork/nex-go/v2`
- `github.com/PretendoNetwork/nex-protocols-common-go/v2`

Both projects are distributed under the GNU Affero General Public License,
version 3. Their license texts are retained in their respective
`third_party/*/LICENSE` files. The vendored copies contain compatibility and
PRUDP changes used by this server; they are selected by the `replace`
directives in the root `go.mod`.

The project also uses `github.com/PretendoNetwork/nex-protocols-go/v2` and the
other modules recorded in `go.mod` and `go.sum`. Their respective upstream
licenses continue to apply.

No Nintendo or SEGA game executable, asset, SDK, certificate, credential, or
private server data is included in this repository.
